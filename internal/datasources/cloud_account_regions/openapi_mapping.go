package cloud_account_regions //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operation covered by
// data.nc2_cloud_account_regions. Shares the `index` operation id
// with the nc2_cloud_account_region managed resource (which uses it
// for Read).
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.RegionController.index", TerraformOp: "data.nc2_cloud_account_regions.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
