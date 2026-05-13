package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_US5_ClusterActions_Smoke exercises the 7 cluster-scoped
// actions against a sandbox cluster (FR-025).
func TestAcc_US5_ClusterActions_Smoke(t *testing.T) {
	SkipIfNotAcc(t)
	envs := SkipIfMissing(t,
		"NC2_API_KEY", "NC2_KEY_ID", "NC2_ISSUER",
		"NC2_TEST_CLUSTER_ID",
		"NC2_TEST_HOST_ID",
	)
	clusterID, hostID := envs[3], envs[4]

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: us5ActionsConfig(clusterID, hostID)},
		},
	})
}

func us5ActionsConfig(clusterID, hostID string) string {
	return fmt.Sprintf(`
action "nc2_cluster_open_support_tunnel" "open" { config { cluster_id = %[1]q } }
action "nc2_cluster_extend_support_tunnel" "extend" { config { cluster_id = %[1]q, duration_hours = 1 } }
action "nc2_cluster_close_support_tunnel" "close" { config { cluster_id = %[1]q } }
action "nc2_cluster_scale_flow_gateway" "scale" { config { cluster_id = %[1]q, target_node_count = 3 } }
action "nc2_cluster_upgrade_flow_gateway" "upgrade" { config { cluster_id = %[1]q } }
action "nc2_cluster_start_recovery" "recover" { config { cluster_id = %[1]q } }
action "nc2_cluster_condemn_host" "condemn" { config { cluster_id = %[1]q, host_id = %[2]q } }
`, clusterID, hostID)
}
