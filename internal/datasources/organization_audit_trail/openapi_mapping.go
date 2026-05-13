package organization_audit_trail //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operation covered by
// data.nc2_organization_audit_trail.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.OrganizationController.list_audit_trails", TerraformOp: "data.nc2_organization_audit_trail.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
