package cluster_start_recovery //nolint:revive,staticcheck

import "testing"

// TestOperationMappings covers FR-021 coverage.
func TestOperationMappings(t *testing.T) {
	t.Parallel()
	if len(OperationMappings) != 1 ||
		OperationMappings[0].OperationID != "CPanelWeb.Api.ClusterController.start_recovery" {
		t.Errorf("got %+v", OperationMappings)
	}
}
