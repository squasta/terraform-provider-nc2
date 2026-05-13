package azure_cluster //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operations covered by
// the Azure cluster resource. Excludes update_access_policy
// (AWS-only per FR-010a).
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.ClusterController.create_azure", TerraformOp: "nc2_azure_cluster.Create"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
