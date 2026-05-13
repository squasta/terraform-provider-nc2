package aws_cluster //nolint:revive,staticcheck

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

// OperationMappings declares the NC2 OpenAPI operations covered by
// the AWS cluster resource. PUT (`update`) is intentionally not
// listed (CI-only ReplaceAll path). Hibernate / resume land here
// because `desired_state` flips route to those endpoints.
//
// CPanelWeb.Api.ClusterController.update / show / index are shared
// with the Azure / GCP clusters and the cluster data sources;
// coverage-check dedups by id.
var OperationMappings = []oapi.Mapping{
	{OperationID: "CPanelWeb.Api.ClusterController.create_aws", TerraformOp: "nc2_aws_cluster.Create"},
	{OperationID: "CPanelWeb.Api.ClusterController.show", TerraformOp: "nc2_aws_cluster.Read"},
	{OperationID: "CPanelWeb.Api.ClusterController.update", TerraformOp: "nc2_aws_cluster.Update.generic.put_replace_all"},
	{OperationID: "CPanelWeb.Api.ClusterController.update (2)", TerraformOp: "nc2_aws_cluster.Update.generic"},
	{OperationID: "CPanelWeb.Api.ClusterController.update_capacity", TerraformOp: "nc2_aws_cluster.Update.capacity"},
	{OperationID: "CPanelWeb.Api.ClusterController.update_ssh_key", TerraformOp: "nc2_aws_cluster.Update.ssh_key"},
	{OperationID: "CPanelWeb.Api.ClusterController.update_license", TerraformOp: "nc2_aws_cluster.Update.license"},
	{OperationID: "CPanelWeb.Api.ClusterController.update_resource_tags", TerraformOp: "nc2_aws_cluster.Update.resource_tags"},
	{OperationID: "CPanelWeb.Api.ClusterController.update_access_policy", TerraformOp: "nc2_aws_cluster.Update.access_policy"},
	{OperationID: "CPanelWeb.Api.ClusterController.terminate", TerraformOp: "nc2_aws_cluster.Delete"},
	{OperationID: "CPanelWeb.Api.ClusterController.hibernate", TerraformOp: "nc2_aws_cluster.Update.hibernate"},
	{OperationID: "CPanelWeb.Api.ClusterController.resume", TerraformOp: "nc2_aws_cluster.Update.resume"},
}

//nolint:gochecknoinits // init is the documented mechanism per R-09.
func init() {
	oapi.Default.Register(OperationMappings...)
}
