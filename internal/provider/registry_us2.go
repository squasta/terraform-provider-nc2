package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	dscluster "github.com/nutanix/terraform-provider-nc2/internal/datasources/cluster"
	dsclusters "github.com/nutanix/terraform-provider-nc2/internal/datasources/clusters"
	rsawscluster "github.com/nutanix/terraform-provider-nc2/internal/resources/aws_cluster"
	rsazurecluster "github.com/nutanix/terraform-provider-nc2/internal/resources/azure_cluster"
	rsgcpcluster "github.com/nutanix/terraform-provider-nc2/internal/resources/gcp_cluster"
)

//nolint:gochecknoinits // init is the documented per-story registration mechanism.
func init() {
	resourceRegistry = append(resourceRegistry,
		func() resource.Resource { return rsawscluster.NewResource() },
		func() resource.Resource { return rsazurecluster.NewResource() },
		func() resource.Resource { return rsgcpcluster.NewResource() },
	)
	dataSourceRegistry = append(dataSourceRegistry,
		func() datasource.DataSource { return dsclusters.NewDataSource() },
		func() datasource.DataSource { return dscluster.NewDataSource() },
	)
}
