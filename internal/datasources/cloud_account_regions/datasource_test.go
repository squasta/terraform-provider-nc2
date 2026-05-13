package cloud_account_regions //nolint:revive,staticcheck

import "testing"

func TestOperationMappings(t *testing.T) {
	t.Parallel()

	if len(OperationMappings) != 1 || OperationMappings[0].OperationID != "CPanelWeb.Api.RegionController.index" {
		t.Errorf("got %+v", OperationMappings)
	}
}
