package aws_cluster //nolint:revive,staticcheck

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
)

// TestRouteUpdate_AllFieldsChange covers the full FR-010 routing
// order: license → ssh_key → capacity → resource_tags →
// access_policy → generic_patch. desired_state is unchanged so
// hibernate/resume is not added.
func TestRouteUpdate_AllFieldsChange(t *testing.T) {
	t.Parallel()

	state := nullModel()
	state.License = types.StringValue("aos")
	state.AOSVersion = types.StringValue("6.7")
	state.SoftwareTier = types.StringValue("pro")
	state.HostAccessSSHKey = types.StringValue("k1")
	state.UseCase = types.StringValue("general")
	state.Capacity = listOfStringMaps(t, []map[string]string{{"host_type": "m5d.metal", "number_of_hosts": "3"}})
	state.ResourceTags = mapOfStrings(t, map[string]string{"env": "demo"})
	state.AccessPolicy = mapOfStrings(t, map[string]string{"mode": "open"})

	plan := nullModel()
	plan.License = types.StringValue("aos")
	plan.AOSVersion = types.StringValue("6.8")
	plan.SoftwareTier = types.StringValue("pro")
	plan.HostAccessSSHKey = types.StringValue("k2")
	plan.UseCase = types.StringValue("production")
	plan.Capacity = listOfStringMaps(t, []map[string]string{{"host_type": "m5d.metal", "number_of_hosts": "5"}})
	plan.ResourceTags = mapOfStrings(t, map[string]string{"env": "prod"})
	plan.AccessPolicy = mapOfStrings(t, map[string]string{"mode": "restricted"})

	ops, err := clustershared.RouteUpdate(computeDiff(state, plan))
	if err != nil {
		t.Fatalf("RouteUpdate: %v", err)
	}
	want := []clustershared.OperationKind{
		clustershared.OpUpdateLicense,
		clustershared.OpUpdateSSHKey,
		clustershared.OpUpdateCapacity,
		clustershared.OpUpdateResourceTags,
		clustershared.OpUpdateAccessPolicy,
		clustershared.OpGenericPatch,
	}
	if len(ops) != len(want) {
		t.Fatalf("ops = %+v; want %+v", ops, want)
	}
	for i, op := range ops {
		if op.Kind != want[i] {
			t.Errorf("ops[%d] = %s; want %s", i, op.Kind, want[i])
		}
	}
}

// TestRouteUpdate_SingleFieldChange_License pins the simplest
// per-attribute routing case.
func TestRouteUpdate_SingleFieldChange_License(t *testing.T) {
	t.Parallel()

	state := nullModel()
	state.License = types.StringValue("aos")
	state.AOSVersion = types.StringValue("6.7")
	state.SoftwareTier = types.StringValue("pro")

	plan := nullModel()
	plan.License = types.StringValue("aos")
	plan.AOSVersion = types.StringValue("6.8")
	plan.SoftwareTier = types.StringValue("pro")

	ops, err := clustershared.RouteUpdate(computeDiff(state, plan))
	if err != nil {
		t.Fatalf("RouteUpdate: %v", err)
	}
	if len(ops) != 1 || ops[0].Kind != clustershared.OpUpdateLicense {
		t.Errorf("ops = %+v; want single OpUpdateLicense", ops)
	}
}

// TestBuildUpdateBodies_LicensePayload pins the body shape sent to
// /clusters/{id}/update-license.
func TestBuildUpdateBodies_LicensePayload(t *testing.T) {
	t.Parallel()

	state := nullModel()
	plan := nullModel()
	plan.License = types.StringValue("aos")
	plan.AOSVersion = types.StringValue("6.8")
	plan.SoftwareTier = types.StringValue("pro")

	bodies := buildUpdateBodies(state, plan)
	if bodies.License["license"] != "aos" || bodies.License["aos_version"] != "6.8" || bodies.License["software_tier"] != "pro" {
		t.Errorf("License body = %+v", bodies.License)
	}
}

// listOfStringMaps is a test helper wrapping types.ListValue.
func listOfStringMaps(t *testing.T, items []map[string]string) types.List {
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

// mapOfStrings is a test helper wrapping types.MapValue.
func mapOfStrings(t *testing.T, in map[string]string) types.Map {
	t.Helper()
	raw := make(map[string]attr.Value, len(in))
	for k, v := range in {
		raw[k] = types.StringValue(v)
	}
	mv, diag := types.MapValue(types.StringType, raw)
	if diag.HasError() {
		t.Fatalf("MapValue: %v", diag)
	}
	return mv
}
