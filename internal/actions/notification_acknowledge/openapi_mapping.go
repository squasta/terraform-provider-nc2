package notification_acknowledge //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares NC2 OpenAPI operation coverage. The
// PATCH variant is the documented "(2)" suffix duplicate (parallels
// the cluster / organization PATCH endpoints, see FR-010b).
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.NotificationController.update (2)", TerraformOp: "nc2_notification_acknowledge.Invoke"},
	{OperationID: "CPanelWeb.Api.NotificationController.update", TerraformOp: "nc2_notification_acknowledge.Invoke.put_replace_all"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
