package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"

	availabilityzones "github.com/nutanix/terraform-provider-nc2/internal/datasources/availability_zones"
	clustercloudresources "github.com/nutanix/terraform-provider-nc2/internal/datasources/cluster_cloud_resources"
	dsnotifications "github.com/nutanix/terraform-provider-nc2/internal/datasources/notifications"
	prismcentrals "github.com/nutanix/terraform-provider-nc2/internal/datasources/prism_centrals"
	remotestorageprofiles "github.com/nutanix/terraform-provider-nc2/internal/datasources/remote_storage_profiles"
	sshkeys "github.com/nutanix/terraform-provider-nc2/internal/datasources/ssh_keys"
	dstask "github.com/nutanix/terraform-provider-nc2/internal/datasources/task"
	dstasks "github.com/nutanix/terraform-provider-nc2/internal/datasources/tasks"
	dsvnets "github.com/nutanix/terraform-provider-nc2/internal/datasources/vnets"
	dsvpcs "github.com/nutanix/terraform-provider-nc2/internal/datasources/vpcs"
)

//nolint:gochecknoinits // init is the documented per-story registration mechanism.
func init() {
	dataSourceRegistry = append(dataSourceRegistry,
		func() datasource.DataSource { return availabilityzones.NewDataSource() },
		func() datasource.DataSource { return sshkeys.NewDataSource() },
		func() datasource.DataSource { return prismcentrals.NewDataSource() },
		func() datasource.DataSource { return dsvnets.NewDataSource() },
		func() datasource.DataSource { return dsvpcs.NewDataSource() },
		func() datasource.DataSource { return remotestorageprofiles.NewDataSource() },
		func() datasource.DataSource { return clustercloudresources.NewDataSource() },
		func() datasource.DataSource { return dsnotifications.NewDataSource() },
		func() datasource.DataSource { return dstasks.NewDataSource() },
		func() datasource.DataSource { return dstask.NewDataSource() },
	)
}
