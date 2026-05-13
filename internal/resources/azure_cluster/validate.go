package azure_cluster //nolint:revive,staticcheck

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Allowed bare-metal SKUs for NC2 on Azure clusters. NC2 only
// supports these two SKUs as cluster hosts on Azure (Nutanix bare
// metal "AN" series); Standard_D / Standard_E and other generic
// Azure VM SKUs are not provisionable as NC2 cluster hosts.
const (
	HostTypeAN36P = "AN36P"
	HostTypeAN64  = "AN64"
)

// AllowedHostTypes returns the closed set of `host_type` values
// permitted in `capacity[*].host_type` on `nc2_azure_cluster`.
// Returned slice is freshly allocated and sorted; callers may
// mutate without affecting later invocations.
func AllowedHostTypes() []string {
	out := []string{HostTypeAN36P, HostTypeAN64}
	sort.Strings(out)
	return out
}

// allowedHostTypeSet is the runtime lookup table backing
// validateCapacityHostTypes; keep in sync with AllowedHostTypes.
var allowedHostTypeSet = map[string]struct{}{
	HostTypeAN36P: {},
	HostTypeAN64:  {},
}

// validateCapacityHostTypes walks the planned `capacity` list and
// emits one error diagnostic per element whose `host_type` is not
// in AllowedHostTypes(). It is a pure function: no I/O, no client
// dependency, fully exercisable from a unit test.
//
// Empty / null / unknown capacity is treated as "no validation
// needed": the framework's Required-attribute machinery already
// enforces presence at the schema level, and unknown values cannot
// be inspected at plan time.
//
// Each error is attached to the precise nested attribute path
// (`capacity[i].host_type`) so the operator gets actionable
// feedback in the plan output.
func validateCapacityHostTypes(capacity types.List) diag.Diagnostics {
	var diags diag.Diagnostics
	if capacity.IsNull() || capacity.IsUnknown() {
		return diags
	}
	allowed := strings.Join(AllowedHostTypes(), ", ")
	for i, raw := range capacity.Elements() {
		hostMap, ok := raw.(types.Map)
		if !ok || hostMap.IsNull() || hostMap.IsUnknown() {
			continue
		}
		htRaw, present := hostMap.Elements()["host_type"]
		if !present {
			continue
		}
		htStr, ok := htRaw.(types.String)
		if !ok || htStr.IsNull() || htStr.IsUnknown() {
			continue
		}
		v := htStr.ValueString()
		if _, ok := allowedHostTypeSet[v]; ok {
			continue
		}
		diags.AddAttributeError(
			path.Root("capacity").AtListIndex(i).AtMapKey("host_type"),
			"unsupported host_type for nc2_azure_cluster",
			fmt.Sprintf("host_type %q is not supported on NC2 on Azure; allowed values: %s", v, allowed),
		)
	}
	return diags
}
