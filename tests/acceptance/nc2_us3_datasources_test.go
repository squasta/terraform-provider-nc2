package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_US3_InventoryDataSources_Smoke exercises the 10 US3 read-
// only data sources against a sandbox tenant (FR-024).
func TestAcc_US3_InventoryDataSources_Smoke(t *testing.T) {
	SkipIfNotAcc(t)
	envs := SkipIfMissing(t,
		"NC2_API_KEY", "NC2_KEY_ID", "NC2_ISSUER",
		"NC2_TEST_CLOUD_ACCOUNT_ID",
		"NC2_TEST_REGION_ID",
		"NC2_TEST_CLUSTER_ID",
		"NC2_TEST_TASK_ID",
	)
	cloudAccountID, regionID, clusterID, taskID := envs[3], envs[4], envs[5], envs[6]

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: us3DataSourcesConfig(cloudAccountID, regionID, clusterID, taskID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.nc2_availability_zones.az", "cloud_account_id"),
				),
			},
		},
	})
}

func us3DataSourcesConfig(cloudAccountID, regionID, clusterID, taskID string) string {
	return fmt.Sprintf(`
data "nc2_availability_zones" "az" {
  cloud_account_id = %q
  region_id        = %q
}

data "nc2_ssh_keys" "ssh" {
  cloud_account_id = %q
  region_id        = %q
}

data "nc2_prism_centrals" "pc" {
  cloud_account_id = %q
  region_id        = %q
}

data "nc2_vnets" "vn" {
  cloud_account_id = %q
  region_id        = %q
}

data "nc2_vpcs" "vp" {
  cloud_account_id = %q
  region_id        = %q
}

data "nc2_remote_storage_profiles" "rsp" {
  cloud_account_id = %q
  region_id        = %q
}

data "nc2_cluster_cloud_resources" "ccr" {
  cluster_id = %q
}

data "nc2_notifications" "n" {}
data "nc2_tasks" "ts" {}
data "nc2_task" "t" { id = %q }
`,
		cloudAccountID, regionID,
		cloudAccountID, regionID,
		cloudAccountID, regionID,
		cloudAccountID, regionID,
		cloudAccountID, regionID,
		cloudAccountID, regionID,
		clusterID,
		taskID,
	)
}
