package cloud_account //nolint:revive,staticcheck

import "testing"

func TestOperationMappings(t *testing.T) {
	t.Parallel()

	if len(OperationMappings) != 1 || OperationMappings[0].OperationID != "CPanelWeb.Api.CloudAccountController.show" {
		t.Errorf("got %+v", OperationMappings)
	}
}
