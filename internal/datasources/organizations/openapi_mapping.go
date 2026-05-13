package organizations

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operation covered by
// data.nc2_organizations.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.OrganizationController.index", TerraformOp: "data.nc2_organizations.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
