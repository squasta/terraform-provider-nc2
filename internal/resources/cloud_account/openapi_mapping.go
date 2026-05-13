package cloud_account //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operations covered by
// the nc2_cloud_account resource. The PUT (`update`) variant is
// listed for coverage purposes only; the resource always PATCHes
// per FR-010b.
//
// `list_cloud_accounts` belongs to data.nc2_cloud_accounts and is
// registered there.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.OrganizationController.create_cloud_account", TerraformOp: "nc2_cloud_account.Create"},
	{OperationID: "CPanelWeb.Api.CloudAccountController.show", TerraformOp: "nc2_cloud_account.Read"},
	{OperationID: "CPanelWeb.Api.CloudAccountController.update (2)", TerraformOp: "nc2_cloud_account.Update"},
	{OperationID: "CPanelWeb.Api.CloudAccountController.update", TerraformOp: "nc2_cloud_account.Update.put_replace_all"},
	{OperationID: "CPanelWeb.Api.CloudAccountController.update_credentials", TerraformOp: "nc2_cloud_account.Update.credentials"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
