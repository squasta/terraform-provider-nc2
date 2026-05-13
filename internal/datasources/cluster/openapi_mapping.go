package cluster

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares NC2 OpenAPI operation coverage for the
// single-cluster data source. Shares `show` with the managed
// resource read; coverage-check dedups by id.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.ClusterController.show", TerraformOp: "data.nc2_cluster.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
