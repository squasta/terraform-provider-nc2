package cloud_account //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operation covered by
// data.nc2_cloud_account. Shares the `show` operation id with the
// nc2_cloud_account managed resource; coverage-check dedups by id.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.CloudAccountController.show", TerraformOp: "data.nc2_cloud_account.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
