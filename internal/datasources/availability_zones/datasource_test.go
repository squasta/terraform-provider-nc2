package availability_zones //nolint:revive,staticcheck

import "testing"

// TestOperationMappings covers FR-021 coverage.
func TestOperationMappings(t *testing.T) {
	t.Parallel()
	if len(OperationMappings) != 1 || OperationMappings[0].OperationID != "CPanelWeb.Api.AvailabilityZoneController.index" {
		t.Errorf("got %+v", OperationMappings)
	}
}
