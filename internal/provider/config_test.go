package provider

import (
	"context"
	"testing"
)

// TestResolve_PrecedenceBlockOverEnv covers the FR-001a rule
// "block > env > file".
func TestResolve_PrecedenceBlockOverEnv(t *testing.T) {
	t.Parallel()

	pcfg, diags := Resolve(context.Background(),
		ProviderBlock{APIKey: "block-ak", KeyID: "block-kid", Issuer: "block-iss"},
		Env{APIKey: "env-ak", KeyID: "env-kid", Issuer: "env-iss"},
		File{Profiles: map[string]FileProfile{"default": {APIKey: "file-ak", KeyID: "file-kid", Issuer: "file-iss"}}},
	)
	if pcfg.Credentials.APIKey != "block-ak" {
		t.Errorf("APIKey = %q; want block-ak", pcfg.Credentials.APIKey)
	}
	if !diags.HasError() && len(diags) == 0 {
		t.Errorf("expected at least one warning about source disagreement")
	}
}

// TestResolve_FallsThroughToFile when block + env are empty.
func TestResolve_FallsThroughToFile(t *testing.T) {
	t.Parallel()

	pcfg, diags := Resolve(context.Background(),
		ProviderBlock{},
		Env{},
		File{Profiles: map[string]FileProfile{"default": {APIKey: "f-ak", KeyID: "f-kid", Issuer: "f-iss"}}},
	)
	if pcfg.Credentials.APIKey != "f-ak" {
		t.Errorf("APIKey = %q", pcfg.Credentials.APIKey)
	}
	if diags.HasError() {
		t.Errorf("unexpected error diags: %v", diags)
	}
}

// TestResolve_NoConflictWarningWhenSourcesAgree.
func TestResolve_NoWarningWhenSourcesAgree(t *testing.T) {
	t.Parallel()

	_, diags := Resolve(context.Background(),
		ProviderBlock{APIKey: "ak", KeyID: "kid", Issuer: "iss"},
		Env{APIKey: "ak", KeyID: "kid", Issuer: "iss"},
		File{Profiles: map[string]FileProfile{"default": {APIKey: "ak", KeyID: "kid", Issuer: "iss"}}},
	)
	for _, d := range diags {
		if d.Severity().String() == "Warning" {
			t.Errorf("unexpected warning when sources agree: %v", d)
		}
	}
}

// TestResolve_MissingCredentialsErrors when nothing is supplied.
func TestResolve_MissingCredentialsErrors(t *testing.T) {
	t.Parallel()

	_, diags := Resolve(context.Background(), ProviderBlock{}, Env{}, File{})
	if !diags.HasError() {
		t.Errorf("expected error diagnostic for missing credentials")
	}
}

// TestParseINI parses [profile] sections + key=value pairs.
func TestParseINI(t *testing.T) {
	t.Parallel()
	data := []byte(`
# header
[default]
api_key = AK1
key_id  = KID1
issuer  = ISS1

[staging]
api_key = AK2
key_id  = KID2
issuer  = ISS2
`)
	f, err := parseINI(data)
	if err != nil {
		t.Fatalf("parseINI: %v", err)
	}
	if f.Profiles["default"].APIKey != "AK1" || f.Profiles["staging"].KeyID != "KID2" {
		t.Errorf("got %+v", f)
	}
}

// TestProfileSelection picks the named profile.
func TestProfileSelection(t *testing.T) {
	t.Parallel()
	pcfg, diags := Resolve(context.Background(),
		ProviderBlock{Profile: "staging"},
		Env{},
		File{Profiles: map[string]FileProfile{
			"default": {APIKey: "default-ak", KeyID: "default-kid", Issuer: "default-iss"},
			"staging": {APIKey: "staging-ak", KeyID: "staging-kid", Issuer: "staging-iss"},
		}},
	)
	if diags.HasError() {
		t.Fatalf("error: %v", diags)
	}
	if pcfg.Credentials.APIKey != "staging-ak" {
		t.Errorf("expected staging profile; got %q", pcfg.Credentials.APIKey)
	}
}
