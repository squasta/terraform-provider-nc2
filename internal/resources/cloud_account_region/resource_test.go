package cloud_account_region //nolint:revive,staticcheck

import "testing"

func TestSensitive_Empty(t *testing.T) {
	t.Parallel()

	if got := Sensitive(); len(got) != 0 {
		t.Errorf("Sensitive() = %v; want empty", got)
	}
}

func TestOperationMappings_CoversRegionOperations(t *testing.T) {
	t.Parallel()

	want := map[string]bool{
		"CPanelWeb.Api.RegionController.create": false,
		"CPanelWeb.Api.RegionController.index":  false,
	}
	for _, m := range OperationMappings {
		if _, ok := want[m.OperationID]; ok {
			want[m.OperationID] = true
		}
	}
	for k, ok := range want {
		if !ok {
			t.Errorf("OperationMappings missing %q", k)
		}
	}
}

func TestSplitImportID(t *testing.T) {
	t.Parallel()

	cases := map[string]int{
		"ca-uuid/region-uuid": 2,
		"only-one":            1,
		"":                    0,
	}
	for in, want := range cases {
		got := splitImportID(in)
		if len(got) != want {
			t.Errorf("split(%q) len = %d; want %d", in, len(got), want)
		}
	}
}

func TestFindRegionByID(t *testing.T) {
	t.Parallel()

	body := map[string]any{
		"data": []any{
			map[string]any{"id": "r1", "region": "us-east-1"},
			map[string]any{"id": "r2", "region": "us-west-2"},
		},
	}
	row, ok := findRegionByID(body, "r2")
	if !ok || row["region"] != "us-west-2" {
		t.Errorf("did not find r2; row=%v ok=%v", row, ok)
	}
	if _, ok := findRegionByID(body, "missing"); ok {
		t.Errorf("expected miss for unknown id")
	}
}

func TestDecodeInto_FromObject(t *testing.T) {
	t.Parallel()

	body := map[string]any{"data": map[string]any{"id": "r-uuid", "region": "us-east-1", "state": "available"}}
	var m model
	if diags := decodeInto(body, &m); diags.HasError() {
		t.Fatalf("decodeInto: %v", diags)
	}
	if m.ID.ValueString() != "r-uuid" || m.Region.ValueString() != "us-east-1" || m.State.ValueString() != "available" {
		t.Errorf("got %+v", m)
	}
}

func TestDecodeInto_FromListPicksFirst(t *testing.T) {
	t.Parallel()

	body := map[string]any{
		"data": []any{
			map[string]any{"id": "r-uuid", "region": "us-east-1", "state": "available"},
		},
	}
	var m model
	if diags := decodeInto(body, &m); diags.HasError() {
		t.Fatalf("decodeInto: %v", diags)
	}
	if m.ID.ValueString() != "r-uuid" {
		t.Errorf("got %+v", m)
	}
}
