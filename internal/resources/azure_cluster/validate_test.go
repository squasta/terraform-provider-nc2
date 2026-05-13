package azure_cluster //nolint:revive,staticcheck

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// capacityList is a small test helper that builds a
// `List<Map<String,String>>` matching the framework type the
// `capacity` attribute uses on Azure clusters.
func capacityList(t *testing.T, items []map[string]string) types.List {
	t.Helper()
	mt := types.MapType{ElemType: types.StringType}
	values := make([]attr.Value, 0, len(items))
	for _, item := range items {
		raw := make(map[string]attr.Value, len(item))
		for k, v := range item {
			raw[k] = types.StringValue(v)
		}
		mv, diag := types.MapValue(types.StringType, raw)
		if diag.HasError() {
			t.Fatalf("MapValue: %v", diag)
		}
		values = append(values, mv)
	}
	lv, diag := types.ListValue(mt, values)
	if diag.HasError() {
		t.Fatalf("ListValue: %v", diag)
	}
	return lv
}

// TestAllowedHostTypes pins the closed set published to operators.
// Adding a new host type requires updating this test deliberately.
func TestAllowedHostTypes(t *testing.T) {
	t.Parallel()

	got := AllowedHostTypes()
	want := []string{"AN36P", "AN64"}
	if len(got) != len(want) {
		t.Fatalf("AllowedHostTypes() = %v; want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("AllowedHostTypes()[%d] = %q; want %q", i, got[i], want[i])
		}
	}
}

// TestValidateCapacityHostTypes_AcceptsAllowedSKUs covers the
// happy path for both supported bare-metal SKUs.
func TestValidateCapacityHostTypes_AcceptsAllowedSKUs(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": HostTypeAN36P, "number_of_hosts": "3"},
		{"host_type": HostTypeAN64, "number_of_hosts": "5"},
	})
	if d := validateCapacityHostTypes(cap); d.HasError() {
		t.Errorf("expected no errors for allowed SKUs; got %v", d)
	}
}

// TestValidateCapacityHostTypes_RejectsAWSSKU pins the most likely
// operator mistake: copy-pasting an AWS SKU into an Azure capacity
// block.
func TestValidateCapacityHostTypes_RejectsAWSSKU(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "m5d.metal", "number_of_hosts": "3"},
	})
	d := validateCapacityHostTypes(cap)
	if !d.HasError() {
		t.Fatalf("expected error for AWS SKU; got %v", d)
	}
	if got := d[0].Detail(); !strings.Contains(got, "m5d.metal") || !strings.Contains(got, "AN36P") || !strings.Contains(got, "AN64") {
		t.Errorf("error detail %q must name the offending value AND list both allowed SKUs", got)
	}
}

// TestValidateCapacityHostTypes_RejectsAzureGenericVM pins the
// other common mistake: using a generic Azure VM SKU
// (Standard_D32s_v4) that NC2 does not provision as a cluster
// host.
func TestValidateCapacityHostTypes_RejectsAzureGenericVM(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "Standard_D32s_v4", "number_of_hosts": "3"},
	})
	d := validateCapacityHostTypes(cap)
	if !d.HasError() {
		t.Fatalf("expected error for Standard_D32s_v4; got %v", d)
	}
}

// TestValidateCapacityHostTypes_PerElementErrors ensures the
// validator emits one diagnostic per offending list element so the
// operator sees every problem in a single plan.
func TestValidateCapacityHostTypes_PerElementErrors(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": HostTypeAN36P, "number_of_hosts": "1"}, // ok
		{"host_type": "m5d.metal", "number_of_hosts": "1"},   // bad
		{"host_type": "n2-standard-32", "number_of_hosts": "1"}, // bad
	})
	d := validateCapacityHostTypes(cap)
	if got := len(d); got != 2 {
		t.Errorf("expected 2 error diagnostics; got %d (%v)", got, d)
	}
}

// TestValidateCapacityHostTypes_NullCapacity_NoError treats a
// null/unknown capacity as "skip"; framework-level Required
// enforcement handles presence.
func TestValidateCapacityHostTypes_NullCapacity_NoError(t *testing.T) {
	t.Parallel()

	if d := validateCapacityHostTypes(types.ListNull(types.MapType{ElemType: types.StringType})); d.HasError() {
		t.Errorf("null capacity must not produce diagnostics; got %v", d)
	}
}

// TestValidateCapacityHostTypes_MissingHostTypeKey_Skipped models
// the case where the operator omits host_type entirely; the
// host-type validator stays silent (other validation surfaces
// would catch it, but that's not this validator's job).
func TestValidateCapacityHostTypes_MissingHostTypeKey_Skipped(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"number_of_hosts": "3"},
	})
	if d := validateCapacityHostTypes(cap); d.HasError() {
		t.Errorf("missing host_type must not produce host-type errors; got %v", d)
	}
}
