package clusters

import "testing"

// TestExtractList covers the canonical body decoder.
func TestExtractList(t *testing.T) {
	t.Parallel()

	got := extractList(map[string]any{
		"data": []any{
			map[string]any{"id": "c1", "cloud_provider": "aws"},
			map[string]any{"id": "c2", "cloud_provider": "azure"},
		},
	})
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
}

// TestOperationMappings covers FR-021 coverage.
func TestOperationMappings(t *testing.T) {
	t.Parallel()

	if len(OperationMappings) != 1 || OperationMappings[0].OperationID != "CPanelWeb.Api.ClusterController.index" {
		t.Errorf("got %+v", OperationMappings)
	}
}
