package notifications

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares NC2 OpenAPI operation coverage.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.NotificationController.index", TerraformOp: "data.nc2_notifications.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
