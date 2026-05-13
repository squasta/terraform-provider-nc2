package vpcs

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares NC2 OpenAPI operation coverage.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.VpcController.index", TerraformOp: "data.nc2_vpcs.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
