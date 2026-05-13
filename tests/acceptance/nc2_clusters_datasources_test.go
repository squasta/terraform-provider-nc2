package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_US2_ClusterDataSources_Smoke exercises the US2 data
// sources against a sandbox tenant (FR-024).
func TestAcc_US2_ClusterDataSources_Smoke(t *testing.T) {
	SkipIfNotAcc(t)
	envs := SkipIfMissing(t,
		"NC2_API_KEY", "NC2_KEY_ID", "NC2_ISSUER",
		"NC2_TEST_CLUSTER_ID",
	)
	clusterID := envs[3]

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: clusterDataSourcesConfig(clusterID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.nc2_cluster.target", "name"),
				),
			},
		},
	})
}

func clusterDataSourcesConfig(clusterID string) string {
	return fmt.Sprintf(`
data "nc2_clusters" "all" {}

data "nc2_cluster" "target" {
  id = %q
}
`, clusterID)
}
