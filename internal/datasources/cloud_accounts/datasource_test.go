package cloud_accounts //nolint:revive,staticcheck

import "testing"

func TestOperationMappings(t *testing.T) {
	t.Parallel()

	if len(OperationMappings) != 1 || OperationMappings[0].OperationID != "CPanelWeb.Api.OrganizationController.list_cloud_accounts" {
		t.Errorf("got %+v", OperationMappings)
	}
}
