package ssh_keys //nolint:revive,staticcheck

import "testing"

// TestOperationMappings covers FR-021 coverage.
func TestOperationMappings(t *testing.T) {
	t.Parallel()
	if len(OperationMappings) != 1 || OperationMappings[0].OperationID != "CPanelWeb.Api.SshKeyController.index" {
		t.Errorf("got %+v", OperationMappings)
	}
}
