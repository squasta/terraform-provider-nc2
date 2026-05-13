package azure_cluster //nolint:revive,staticcheck

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/nutanix/terraform-provider-nc2/internal/resources/clustershared"
)

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
	return body
}

// computeDiff inspects state vs plan and produces the boolean
// per-attribute diff that clustershared.RouteUpdate consumes.
// Hibernate / resume are intentionally absent — `desired_state` is
// AWS-only (FR-012) and is not part of the Azure schema, so the
// returned Diff's DesiredStateFrom/To always remain empty.
func computeDiff(state, plan *model) clustershared.Diff {
	d := clustershared.Diff{
		License: !plan.License.Equal(state.License) ||
			!plan.AOSVersion.Equal(state.AOSVersion) ||
			!plan.SoftwareTier.Equal(state.SoftwareTier),
		SSHKey:       !plan.HostAccessSSHKey.Equal(state.HostAccessSSHKey),
		Capacity:     !plan.Capacity.Equal(state.Capacity),
		ResourceTags: !plan.ResourceTags.Equal(state.ResourceTags),
	}
	if !plan.UseCase.Equal(state.UseCase) {
		d.GenericMutable = true
	}
	return d
}

func buildUpdateBodies(_, plan *model) clustershared.UpdateBodies {
	bodies := clustershared.UpdateBodies{
		License: map[string]any{
			"license":       plan.License.ValueString(),
			"aos_version":   plan.AOSVersion.ValueString(),
			"software_tier": plan.SoftwareTier.ValueString(),
		},
		SSHKey: map[string]any{
			"host_access_ssh_key": plan.HostAccessSSHKey.ValueString(),
		},
		GenericPatch: map[string]any{"use_case": plan.UseCase.ValueString()},
	}
	if cap := mapsListFromTFList(plan.Capacity); cap != nil {
		bodies.Capacity = map[string]any{"capacity": cap}
	}
	if tags := stringMapFromTFMap(plan.ResourceTags); tags != nil {
		bodies.ResourceTags = map[string]any{"resource_tags": tags}
	}
	return bodies
}

func addMap(dst map[string]any, key string, v types.Map) {
	if m := stringMapFromTFMap(v); m != nil {
		dst[key] = m
	}
}

func addCapacity(dst map[string]any, v types.List) {
	if l := mapsListFromTFList(v); l != nil {
		dst["capacity"] = l
	}
}

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
