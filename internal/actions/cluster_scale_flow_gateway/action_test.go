package cluster_scale_flow_gateway //nolint:revive,staticcheck

import "testing"

// TestOperationMappings covers FR-021 coverage.
func TestOperationMappings(t *testing.T) {
	t.Parallel()
	if len(OperationMappings) != 1 ||
		OperationMappings[0].OperationID != "CPanelWeb.Api.ClusterController.scale_out_fgw" {
		t.Errorf("got %+v", OperationMappings)
	}
}
