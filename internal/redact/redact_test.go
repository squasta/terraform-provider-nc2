package redact

import (
	"reflect"
	"testing"
)

// Sentinel used in every test to assert the redacted value.
const want = "(sensitive)"

// TestNewRegistry_EmptyExtra checks the simple constructor case: an empty
// registry, where redaction is driven purely by FR-002 pattern matching.
func TestNewRegistry_EmptyExtra(t *testing.T) {
	t.Parallel()

	r := NewRegistry(nil)
	in := map[string]any{
		"name":         "demo",
		"api_key":      "AKIA_SECRET",
		"description":  "free-text",
		"PASSWORD":     "p@ssw0rd",
		"organization": "org-1",
	}
	out := r.Redact(in)

	expect := map[string]any{
		"name":         "demo",
		"api_key":      want,
		"description":  "free-text",
		"PASSWORD":     want,
		"organization": "org-1",
	}
	if !reflect.DeepEqual(out, expect) {
		t.Errorf("Redact = %v, want %v", out, expect)
	}
}

// TestRedact_DeepCopyIsolation verifies that Redact never mutates its input,
// only returns a new redacted map. This is critical because audit records,
// API responses, and Terraform state values flow through Redact and the
// caller must not see them altered.
func TestRedact_DeepCopyIsolation(t *testing.T) {
	t.Parallel()

	in := map[string]any{
		"api_key": "AKIA",
		"nested":  map[string]any{"secret": "shh"},
	}
	original := map[string]any{
		"api_key": "AKIA",
		"nested":  map[string]any{"secret": "shh"},
	}

	r := NewRegistry(nil)
	_ = r.Redact(in)

	if !reflect.DeepEqual(in, original) {
		t.Errorf("Redact mutated input map: got %v, want %v", in, original)
	}
}

// TestRedact_NestedMaps exercises recursion: redaction must descend into
// nested map[string]any values (the common shape of NC2 JSON response trees).
//
// access_key_id does NOT match any FR-002 name pattern (credential, password,
// secret, token, private_key, api_key) — it has to be classified through the
// per-resource registry instead. The test below registers it explicitly to
// pin the both-mechanisms behavior in a single fixture.
func TestRedact_NestedMaps(t *testing.T) {
	t.Parallel()

	in := map[string]any{
		"id": "uuid-1",
		"credentials": map[string]any{
			"access_key_id":     "AKIA1234",
			"secret_access_key": "shhh",
			"role_arn":          "arn:aws:iam::123:role/demo",
		},
		"meta": map[string]any{
			"description": "ok",
			"token":       "bearer-xyz",
		},
	}
	r := NewRegistry([]string{"credentials.access_key_id"})
	out := r.Redact(in)

	creds, ok := out["credentials"].(map[string]any)
	if !ok {
		t.Fatalf("credentials missing or wrong type: %T", out["credentials"])
	}
	// access_key_id is classified via per-resource registry; secret_access_key
	// matches the FR-002 "secret" pattern; role_arn matches no pattern and is
	// not in the registry, so it must pass through unchanged.
	if creds["access_key_id"] != want {
		t.Errorf("access_key_id = %v, want %v", creds["access_key_id"], want)
	}
	if creds["secret_access_key"] != want {
		t.Errorf("secret_access_key = %v, want %v", creds["secret_access_key"], want)
	}
	if creds["role_arn"] != "arn:aws:iam::123:role/demo" {
		t.Errorf("role_arn changed: %v", creds["role_arn"])
	}

	meta, ok := out["meta"].(map[string]any)
	if !ok {
		t.Fatalf("meta missing")
	}
	if meta["token"] != want {
		t.Errorf("token = %v, want %v", meta["token"], want)
	}
	if meta["description"] != "ok" {
		t.Errorf("description changed: %v", meta["description"])
	}
}

// TestRedact_ListsOfMaps exercises recursion into []any whose elements are
// themselves maps — the NC2 API returns this shape from every list endpoint.
func TestRedact_ListsOfMaps(t *testing.T) {
	t.Parallel()

	in := map[string]any{
		"items": []any{
			map[string]any{"name": "a", "secret": "x"},
			map[string]any{"name": "b", "secret": "y"},
		},
	}
	r := NewRegistry(nil)
	out := r.Redact(in)

	items, ok := out["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("items missing or wrong shape: %v", out["items"])
	}
	for i, raw := range items {
		m, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("items[%d] not a map: %T", i, raw)
		}
		if m["secret"] != want {
			t.Errorf("items[%d].secret = %v, want %v", i, m["secret"], want)
		}
		if m["name"] == want {
			t.Errorf("items[%d].name was redacted (shouldn't be): %v", i, m["name"])
		}
	}
}

// TestRedact_ExtraRegistryEntries verifies that a per-resource registry entry
// for an attribute path NOT matched by FR-002 patterns is still redacted.
// This is the FR-002 "per-resource sensitive registry" mechanism.
//
// Registry entries use dotted attribute paths (consistent with data-model.md).
func TestRedact_ExtraRegistryEntries(t *testing.T) {
	t.Parallel()

	extra := []string{
		"network.prism_element_access_policy.ip_addresses",
	}
	in := map[string]any{
		"network": map[string]any{
			"prism_element_access_policy": map[string]any{
				"ip_addresses": []any{"10.0.0.1", "10.0.0.2"},
				"mode":         "restricted",
			},
		},
	}
	r := NewRegistry(extra)
	out := r.Redact(in)

	pol, _ := out["network"].(map[string]any)["prism_element_access_policy"].(map[string]any)
	if pol["ip_addresses"] != want {
		t.Errorf("registry-marked path not redacted: %v", pol["ip_addresses"])
	}
	if pol["mode"] != "restricted" {
		t.Errorf("non-marked path was redacted: %v", pol["mode"])
	}
}

// TestRedact_Idempotent verifies redacting twice produces the same result.
// A sensitive value is replaced with the sentinel "(sensitive)" — running
// Redact again must not mistake the sentinel for new data.
func TestRedact_Idempotent(t *testing.T) {
	t.Parallel()

	r := NewRegistry(nil)
	in := map[string]any{
		"api_key": "AKIA",
		"meta":    map[string]any{"description": "ok"},
	}
	once := r.Redact(in)
	twice := r.Redact(once)
	if !reflect.DeepEqual(once, twice) {
		t.Errorf("Redact not idempotent:\n once = %v\n twice = %v", once, twice)
	}
}

// TestRedact_MissingRegistryPathNoOp checks that a registry entry pointing to
// a path absent from the input is a silent no-op (no panic, no error).
func TestRedact_MissingRegistryPathNoOp(t *testing.T) {
	t.Parallel()

	r := NewRegistry([]string{"nonexistent.deep.path"})
	in := map[string]any{"name": "demo"}
	out := r.Redact(in)
	if !reflect.DeepEqual(out, in) {
		t.Errorf("Redact mutated unrelated data: %v vs %v", out, in)
	}
}

// TestRedact_NilInputReturnsNil verifies nil-safety.
func TestRedact_NilInputReturnsNil(t *testing.T) {
	t.Parallel()

	r := NewRegistry(nil)
	if got := r.Redact(nil); got != nil {
		t.Errorf("Redact(nil) = %v, want nil", got)
	}
}

// TestRedact_NonMapValueAtPathNoOp verifies that a registry path whose parent
// is not a map (e.g., a scalar where the path expected a sub-object) is a
// silent no-op rather than a panic.
func TestRedact_NonMapValueAtPathNoOp(t *testing.T) {
	t.Parallel()

	r := NewRegistry([]string{"a.b.c"})
	in := map[string]any{"a": "scalar"}
	out := r.Redact(in)
	if out["a"] != "scalar" {
		t.Errorf("scalar at registry path was mutated: %v", out["a"])
	}
}
