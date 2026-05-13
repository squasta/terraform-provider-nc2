package clusters

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares NC2 OpenAPI operation coverage for the
// clusters list data source.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.ClusterController.index", TerraformOp: "data.nc2_clusters.Read"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
