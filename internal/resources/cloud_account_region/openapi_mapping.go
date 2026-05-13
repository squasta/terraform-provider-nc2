package cloud_account_region //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operations covered by
// the nc2_cloud_account_region resource. NC2 has no DELETE endpoint
// for regions; destroy is state-only (FR-007).
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.RegionController.create", TerraformOp: "nc2_cloud_account_region.Create"},
	{OperationID: "CPanelWeb.Api.RegionController.index", TerraformOp: "nc2_cloud_account_region.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
