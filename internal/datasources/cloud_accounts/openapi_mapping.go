package cloud_accounts //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operation covered by
// data.nc2_cloud_accounts.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.OrganizationController.list_cloud_accounts", TerraformOp: "data.nc2_cloud_accounts.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
