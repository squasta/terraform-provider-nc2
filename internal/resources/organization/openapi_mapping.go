package organization

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operations covered by
// the nc2_organization resource. PUT (`update`) is intentionally
// listed so coverage-check sees full path coverage; the Update
// method itself only ever PATCHes.
//
// `index` belongs to data.nc2_organizations and is registered there.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.OrganizationController.create", TerraformOp: "nc2_organization.Create"},
	{OperationID: "CPanelWeb.Api.OrganizationController.show", TerraformOp: "nc2_organization.Read"},
	{OperationID: "CPanelWeb.Api.OrganizationController.update (2)", TerraformOp: "nc2_organization.Update"},
	{OperationID: "CPanelWeb.Api.OrganizationController.update", TerraformOp: "nc2_organization.Update.put_replace_all"},
	{OperationID: "CPanelWeb.Api.OrganizationController.terminate", TerraformOp: "nc2_organization.Delete"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
