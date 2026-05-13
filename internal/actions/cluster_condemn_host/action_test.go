package cluster_condemn_host //nolint:revive,staticcheck

import "testing"

// TestOperationMappings covers FR-021 coverage.
func TestOperationMappings(t *testing.T) {
	t.Parallel()
	if len(OperationMappings) != 1 ||
		OperationMappings[0].OperationID != "CPanelWeb.Api.ClusterController.condemn_host" {
		t.Errorf("got %+v", OperationMappings)
	}
}
