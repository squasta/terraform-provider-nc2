package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	dscloudaccount "github.com/nutanix/terraform-provider-nc2/internal/datasources/cloud_account"
	dscloudaccountregions "github.com/nutanix/terraform-provider-nc2/internal/datasources/cloud_account_regions"
	dscloudaccounts "github.com/nutanix/terraform-provider-nc2/internal/datasources/cloud_accounts"
	dsorganization "github.com/nutanix/terraform-provider-nc2/internal/datasources/organization"
	dsorganizationaudittrail "github.com/nutanix/terraform-provider-nc2/internal/datasources/organization_audit_trail"
	dsorganizations "github.com/nutanix/terraform-provider-nc2/internal/datasources/organizations"
	rscloudaccount "github.com/nutanix/terraform-provider-nc2/internal/resources/cloud_account"
	rscloudaccountregion "github.com/nutanix/terraform-provider-nc2/internal/resources/cloud_account_region"
	rsorganization "github.com/nutanix/terraform-provider-nc2/internal/resources/organization"
)

//nolint:gochecknoinits // init is the documented per-story registration mechanism.
func init() {
	resourceRegistry = append(resourceRegistry,
		func() resource.Resource { return rsorganization.NewResource() },
		func() resource.Resource { return rscloudaccount.NewResource() },
		func() resource.Resource { return rscloudaccountregion.NewResource() },
	)
	dataSourceRegistry = append(dataSourceRegistry,
		func() datasource.DataSource { return dsorganizations.NewDataSource() },
		func() datasource.DataSource { return dsorganization.NewDataSource() },
		func() datasource.DataSource { return dsorganizationaudittrail.NewDataSource() },
		func() datasource.DataSource { return dscloudaccounts.NewDataSource() },
		func() datasource.DataSource { return dscloudaccount.NewDataSource() },
		func() datasource.DataSource { return dscloudaccountregions.NewDataSource() },
	)
}
