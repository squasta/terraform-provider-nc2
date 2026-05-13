package cluster_close_support_tunnel //nolint:revive,staticcheck

import "testing"

// TestOperationMappings covers FR-021 coverage.
func TestOperationMappings(t *testing.T) {
	t.Parallel()
	if len(OperationMappings) != 1 ||
		OperationMappings[0].OperationID != "CPanelWeb.Api.ClusterController.close_support_tunnel" {
		t.Errorf("got %+v", OperationMappings)
	}
}
