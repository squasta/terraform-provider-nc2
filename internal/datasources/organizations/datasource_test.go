package organizations

import "testing"

func TestOperationMappings(t *testing.T) {
	t.Parallel()

	if len(OperationMappings) != 1 || OperationMappings[0].OperationID != "CPanelWeb.Api.OrganizationController.index" {
		t.Errorf("got %+v", OperationMappings)
	}
}
