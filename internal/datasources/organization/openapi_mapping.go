package organization

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operation covered by
// data.nc2_organization. Shares the `show` operation id with the
// nc2_organization managed resource; coverage-check dedups by id.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.OrganizationController.show", TerraformOp: "data.nc2_organization.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
