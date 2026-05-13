package cluster_start_recovery //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares NC2 OpenAPI operation coverage.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.ClusterController.start_recovery", TerraformOp: "nc2_cluster_start_recovery.Invoke"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
