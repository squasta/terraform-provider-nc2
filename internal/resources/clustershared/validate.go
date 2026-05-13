package clustershared

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NC2 cluster sizing rules (apply to every cloud — AWS, Azure, GCP):
//
//   - The smallest supported cluster has exactly 1 host
//     (single-host cluster).
//   - "Production" clusters range from 3 to 28 hosts inclusive.
//   - A 2-host cluster is intentionally rejected by NC2 because it
//     has no quorum / metadata-redundancy story.
//
// These constants drive both the runtime check (ValidateCapacityHostCount)
// and the documentation surface (AllowedClusterHostCounts), so they
// must stay in sync with the NC2 control-plane policy.
const (
	MinClusterHosts           = 1
	MinProductionClusterHosts = 3
	MaxClusterHosts           = 28
)

// AllowedClusterHostCounts returns the closed, sorted set of valid
// totals for `sum(capacity[*].number_of_hosts)` on any cluster
// resource: [1, 3, 4, 5, ..., 28].
//
// The slice is freshly allocated each call; mutations by callers
// do not affect later invocations.
func AllowedClusterHostCounts() []int {
	out := make([]int, 0, 1+(MaxClusterHosts-MinProductionClusterHosts+1))
	out = append(out, MinClusterHosts)
	for n := MinProductionClusterHosts; n <= MaxClusterHosts; n++ {
		out = append(out, n)
	}
	return out
}

// IsAllowedClusterHostCount reports whether n satisfies the cluster
// sizing rule (1, or 3..28 inclusive — never 2, never <1, never
// >28).
func IsAllowedClusterHostCount(n int) bool {
	if n == MinClusterHosts {
		return true
	}
	return n >= MinProductionClusterHosts && n <= MaxClusterHosts
}

// ValidateCapacityHostCount walks the planned `capacity` list and
// returns one or more error diagnostics when:
//
//  1. A `number_of_hosts` element is not parseable as an integer
//     (per-element diagnostic, attached to the offending path).
//  2. A `number_of_hosts` element is < 1 (per-element diagnostic).
//  3. The sum of every fully-known element's `number_of_hosts`
//     fails IsAllowedClusterHostCount (single aggregate diagnostic
//     attached to the `capacity` root).
//
// Empty / null / unknown capacity, or capacity with any unknown
// `number_of_hosts` element, is treated as "skip the aggregate
// check" — the framework's Required-attribute machinery handles
// presence and unknown values cannot be summed pre-apply. The
// per-element checks still run for any element that is fully
// known.
//
// Why a single shared validator? The cluster sizing rule is
// identical for AWS, Azure, and GCP — putting the check here
// means the three resources cannot drift apart, and the unit
// tests below exercise every cloud's behaviour through one
// surface.
func ValidateCapacityHostCount(capacity types.List) diag.Diagnostics {
	var diags diag.Diagnostics
	if capacity.IsNull() || capacity.IsUnknown() {
		return diags
	}

	total := 0
	counted := 0
	anyUnknown := false
	anyParseError := false

	for i, raw := range capacity.Elements() {
		hostMap, ok := raw.(types.Map)
		if !ok || hostMap.IsNull() || hostMap.IsUnknown() {
			anyUnknown = true
			continue
		}
		nohRaw, present := hostMap.Elements()["number_of_hosts"]
		if !present {
			continue
		}
		nohStr, ok := nohRaw.(types.String)
		if !ok || nohStr.IsNull() || nohStr.IsUnknown() {
			anyUnknown = true
			continue
		}
		v := nohStr.ValueString()
		n, err := strconv.Atoi(v)
		if err != nil {
			diags.AddAttributeError(
				path.Root("capacity").AtListIndex(i).AtMapKey("number_of_hosts"),
				"invalid number_of_hosts",
				fmt.Sprintf("number_of_hosts %q must be a positive integer encoded as a string", v),
			)
			anyParseError = true
			continue
		}
		if n < 1 {
			diags.AddAttributeError(
				path.Root("capacity").AtListIndex(i).AtMapKey("number_of_hosts"),
				"invalid number_of_hosts",
				fmt.Sprintf("number_of_hosts %d must be at least 1 per capacity element", n),
			)
			anyParseError = true
			continue
		}
		total += n
		counted++
	}

	// Skip the aggregate check when:
	//   - any element couldn't be inspected at plan time (unknowns)
	//   - any element produced a per-element diagnostic (don't pile on)
	//   - no element actually contributed (e.g. number_of_hosts keys
	//     all missing — that's a different validator's job)
	if anyUnknown || anyParseError || counted == 0 {
		return diags
	}

	if !IsAllowedClusterHostCount(total) {
		diags.AddAttributeError(
			path.Root("capacity"),
			"unsupported cluster host count",
			fmt.Sprintf(
				"sum of capacity[*].number_of_hosts = %d is not a supported NC2 cluster size; "+
					"allowed totals are %d (single-host) or %d through %d inclusive — "+
					"a 2-host cluster has no quorum and is rejected by NC2",
				total, MinClusterHosts, MinProductionClusterHosts, MaxClusterHosts,
			),
		)
	}
	return diags
}
