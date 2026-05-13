package aws_cluster //nolint:revive,staticcheck

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
)

// buildCreateBody flattens the planned model into the request body
// shape expected by POST /clusters/aws.
func buildCreateBody(m *model) map[string]any {
	body := map[string]any{
		"organization_id":     m.OrganizationID.ValueString(),
		"cloud_account_id":    m.CloudAccountID.ValueString(),
		"name":                m.Name.ValueString(),
		"region":              m.Region.ValueString(),
		"host_access_ssh_key": m.HostAccessSSHKey.ValueString(),
		"license":             m.License.ValueString(),
		"aos_version":         m.AOSVersion.ValueString(),
		"software_tier":       m.SoftwareTier.ValueString(),
	}
	if v := m.UseCase.ValueString(); v != "" {
		body["use_case"] = v
	}
	addMap(body, "redundancy", m.Redundancy)
	addMap(body, "network", m.Network)
	addCapacity(body, m.Capacity)
	addMap(body, "resource_tags", m.ResourceTags)
	addMap(body, "access_policy", m.AccessPolicy)
	return body
}

// computeDiff inspects state vs plan and produces the boolean
// per-attribute diff that clustershared.RouteUpdate consumes.
func computeDiff(state, plan *model) clustershared.Diff {
	d := clustershared.Diff{
		License:        !plan.License.Equal(state.License) ||
			!plan.AOSVersion.Equal(state.AOSVersion) ||
			!plan.SoftwareTier.Equal(state.SoftwareTier),
		SSHKey:       !plan.HostAccessSSHKey.Equal(state.HostAccessSSHKey),
		Capacity:     !plan.Capacity.Equal(state.Capacity),
		ResourceTags: !plan.ResourceTags.Equal(state.ResourceTags),
		AccessPolicy: !plan.AccessPolicy.Equal(state.AccessPolicy),
	}
	if !plan.UseCase.Equal(state.UseCase) {
		d.GenericMutable = true
	}
	if state.DesiredState.ValueString() != plan.DesiredState.ValueString() &&
		!plan.DesiredState.IsNull() && !plan.DesiredState.IsUnknown() {
		d.DesiredStateFrom = state.DesiredState.ValueString()
		d.DesiredStateTo = plan.DesiredState.ValueString()
	}
	return d
}

// buildUpdateBodies assembles per-endpoint payloads for every diff
// the router might schedule. Empty payloads are nil to keep the
// generated request body terse.
func buildUpdateBodies(state, plan *model) clustershared.UpdateBodies {
	_ = state
	bodies := clustershared.UpdateBodies{}
	bodies.License = map[string]any{
		"license":       plan.License.ValueString(),
		"aos_version":   plan.AOSVersion.ValueString(),
		"software_tier": plan.SoftwareTier.ValueString(),
	}
	bodies.SSHKey = map[string]any{
		"host_access_ssh_key": plan.HostAccessSSHKey.ValueString(),
	}
	if cap := mapsListFromTFList(plan.Capacity); cap != nil {
		bodies.Capacity = map[string]any{"capacity": cap}
	}
	if tags := stringMapFromTFMap(plan.ResourceTags); tags != nil {
		bodies.ResourceTags = map[string]any{"resource_tags": tags}
	}
	if ap := stringMapFromTFMap(plan.AccessPolicy); ap != nil {
		bodies.AccessPolicy = map[string]any{"access_policy": ap}
	}
	bodies.GenericPatch = map[string]any{"use_case": plan.UseCase.ValueString()}
	bodies.HibernateBody = map[string]any{}
	bodies.ResumeBody = map[string]any{}
	return bodies
}

func addMap(dst map[string]any, key string, v types.Map) {
	m := stringMapFromTFMap(v)
	if m != nil {
		dst[key] = m
	}
}

func addCapacity(dst map[string]any, v types.List) {
	if l := mapsListFromTFList(v); l != nil {
		dst["capacity"] = l
	}
}

// stringMapFromTFMap converts a Map<String,String> to a Go map.
// Returns nil for null / unknown values.
func stringMapFromTFMap(v types.Map) map[string]string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	out := make(map[string]string, len(v.Elements()))
	for k, val := range v.Elements() {
		s, ok := val.(types.String)
		if !ok {
			continue
		}
		out[k] = s.ValueString()
	}
	return out
}

// mapsListFromTFList converts a List<Map<String,String>> to a slice
// of Go maps. Returns nil for null / unknown values.
func mapsListFromTFList(v types.List) []map[string]string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	elems := v.Elements()
	out := make([]map[string]string, 0, len(elems))
	for _, e := range elems {
		m, ok := e.(types.Map)
		if !ok {
			continue
		}
		out = append(out, stringMapFromTFMap(m))
	}
	return out
}

func stringValue(s string) types.String { return types.StringValue(s) }
