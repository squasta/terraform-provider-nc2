package clustershared

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// capacityList is a small test helper that builds a
// `List<Map<String,String>>` matching the framework type the
// `capacity` attribute uses on every cluster resource.
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

// TestAllowedClusterHostCounts pins the closed set published to
// operators. Adding or removing a count requires updating this
// test deliberately so the spec, docs, and runtime stay in sync.
func TestAllowedClusterHostCounts(t *testing.T) {
	t.Parallel()

	got := AllowedClusterHostCounts()
	want := []int{
		1,
		3, 4, 5, 6, 7, 8, 9, 10,
		11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
		21, 22, 23, 24, 25, 26, 27, 28,
	}
	if len(got) != len(want) {
		t.Fatalf("AllowedClusterHostCounts() len = %d; want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("AllowedClusterHostCounts()[%d] = %d; want %d", i, got[i], want[i])
		}
	}
	// Explicitly assert the gap at 2.
	for _, n := range got {
		if n == 2 {
			t.Errorf("AllowedClusterHostCounts must not include 2 (no quorum)")
		}
	}
}

// TestIsAllowedClusterHostCount_Boundaries exercises every edge of
// the rule with one table.
func TestIsAllowedClusterHostCount_Boundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		n    int
		want bool
	}{
		{-5, false},
		{0, false},
		{1, true},
		{2, false},
		{3, true},
		{4, true},
		{15, true},
		{27, true},
		{28, true},
		{29, false},
		{1000, false},
	}
	for _, tc := range cases {
		tc := tc
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := IsAllowedClusterHostCount(tc.n); got != tc.want {
				t.Errorf("IsAllowedClusterHostCount(%d) = %v; want %v", tc.n, got, tc.want)
			}
		})
	}
}

// TestValidateCapacityHostCount_AcceptsSingleHost covers the smallest
// supported cluster size.
func TestValidateCapacityHostCount_AcceptsSingleHost(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P", "number_of_hosts": "1"},
	})
	if d := ValidateCapacityHostCount(cap); d.HasError() {
		t.Errorf("expected no error for 1-host cluster; got %v", d)
	}
}

// TestValidateCapacityHostCount_AcceptsThreeHosts covers the smallest
// production cluster size.
func TestValidateCapacityHostCount_AcceptsThreeHosts(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P", "number_of_hosts": "3"},
	})
	if d := ValidateCapacityHostCount(cap); d.HasError() {
		t.Errorf("expected no error for 3-host cluster; got %v", d)
	}
}

// TestValidateCapacityHostCount_AcceptsMaxHosts covers the upper
// boundary.
func TestValidateCapacityHostCount_AcceptsMaxHosts(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P", "number_of_hosts": "28"},
	})
	if d := ValidateCapacityHostCount(cap); d.HasError() {
		t.Errorf("expected no error for 28-host cluster; got %v", d)
	}
}

// TestValidateCapacityHostCount_RejectsTwoHosts pins the most
// important rule: a 2-host cluster has no quorum and must be
// rejected with a clear, actionable message.
func TestValidateCapacityHostCount_RejectsTwoHosts(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P", "number_of_hosts": "2"},
	})
	d := ValidateCapacityHostCount(cap)
	if !d.HasError() {
		t.Fatalf("expected error for 2-host cluster; got %v", d)
	}
	got := d[0].Detail()
	if !strings.Contains(got, "2") || !strings.Contains(got, "quorum") {
		t.Errorf("error detail %q must name the offending total AND mention quorum", got)
	}
}

// TestValidateCapacityHostCount_RejectsTwoHosts_AcrossGroups
// confirms the rule applies to the SUM, not each element in
// isolation: two host groups with one host each still violates
// the no-2-host rule.
func TestValidateCapacityHostCount_RejectsTwoHosts_AcrossGroups(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P", "number_of_hosts": "1"},
		{"host_type": "AN64", "number_of_hosts": "1"},
	})
	d := ValidateCapacityHostCount(cap)
	if !d.HasError() {
		t.Fatalf("expected error for 1+1=2 host cluster; got %v", d)
	}
}

// TestValidateCapacityHostCount_RejectsZeroHosts covers the
// "missing-N" mistake.
func TestValidateCapacityHostCount_RejectsZeroHosts(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P", "number_of_hosts": "0"},
	})
	d := ValidateCapacityHostCount(cap)
	if !d.HasError() {
		t.Fatalf("expected error for 0-host element; got %v", d)
	}
	if got := d[0].Detail(); !strings.Contains(got, "at least 1") {
		t.Errorf("error detail %q must explain the per-element minimum", got)
	}
}

// TestValidateCapacityHostCount_RejectsAboveMax covers the upper
// boundary the other way.
func TestValidateCapacityHostCount_RejectsAboveMax(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P", "number_of_hosts": "29"},
	})
	if d := ValidateCapacityHostCount(cap); !d.HasError() {
		t.Errorf("expected error for 29-host cluster; got %v", d)
	}
}

// TestValidateCapacityHostCount_RejectsNonInteger pins the
// per-element parse path.
func TestValidateCapacityHostCount_RejectsNonInteger(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P", "number_of_hosts": "three"},
	})
	d := ValidateCapacityHostCount(cap)
	if !d.HasError() {
		t.Fatalf("expected error for non-integer; got %v", d)
	}
	if got := d[0].Detail(); !strings.Contains(got, "three") {
		t.Errorf("error detail %q must include the offending value", got)
	}
}

// TestValidateCapacityHostCount_NullCapacity_NoError treats null
// capacity as "skip" — framework Required-attribute machinery
// catches presence at the schema layer.
func TestValidateCapacityHostCount_NullCapacity_NoError(t *testing.T) {
	t.Parallel()

	if d := ValidateCapacityHostCount(types.ListNull(types.MapType{ElemType: types.StringType})); d.HasError() {
		t.Errorf("null capacity must not produce diagnostics; got %v", d)
	}
}

// TestValidateCapacityHostCount_MissingNumberOfHostsKey_Skipped
// keeps this validator tightly scoped: missing keys are out of
// scope (other validators surface them).
func TestValidateCapacityHostCount_MissingNumberOfHostsKey_Skipped(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P"},
	})
	if d := ValidateCapacityHostCount(cap); d.HasError() {
		t.Errorf("missing number_of_hosts must not produce diagnostics from this validator; got %v", d)
	}
}

// TestValidateCapacityHostCount_AggregateAcrossGroupsAccepted shows
// the typical multi-group case (a 6-host AN36P group + a 4-host
// AN64 group = 10 hosts total) is accepted.
func TestValidateCapacityHostCount_AggregateAcrossGroupsAccepted(t *testing.T) {
	t.Parallel()

	cap := capacityList(t, []map[string]string{
		{"host_type": "AN36P", "number_of_hosts": "6"},
		{"host_type": "AN64", "number_of_hosts": "4"},
	})
	if d := ValidateCapacityHostCount(cap); d.HasError() {
		t.Errorf("expected no error for 6+4=10 host cluster; got %v", d)
	}
}
