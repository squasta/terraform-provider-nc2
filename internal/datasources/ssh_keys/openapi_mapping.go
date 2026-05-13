package ssh_keys //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares NC2 OpenAPI operation coverage.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.SshKeyController.index", TerraformOp: "data.nc2_ssh_keys.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
