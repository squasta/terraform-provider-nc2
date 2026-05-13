package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_GCPCluster_Lifecycle covers FR-023 for nc2_gcp_cluster.
func TestAcc_GCPCluster_Lifecycle(t *testing.T) {
	SkipIfNotAcc(t)
	envs := SkipIfMissing(t,
		"NC2_API_KEY", "NC2_KEY_ID", "NC2_ISSUER",
		"NC2_TEST_ORGANIZATION_ID",
		"NC2_TEST_GCP_CLOUD_ACCOUNT_ID",
		"NC2_TEST_GCP_PROJECT_ID",
		"NC2_TEST_GCP_SSH_KEY_NAME",
	)
	orgID, cloudAccountID, projectID, sshKey := envs[3], envs[4], envs[5], envs[6]
	name := fmt.Sprintf("tf-acc-gcp-%d", randInt())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: gcpClusterConfig(orgID, cloudAccountID, projectID, name, sshKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nc2_gcp_cluster.target", "name", name),
				),
			},
		},
	})
}

func gcpClusterConfig(orgID, cloudAccountID, projectID, name, sshKey string) string {
	return fmt.Sprintf(`
resource "nc2_gcp_cluster" "target" {
  organization_id     = %q
  cloud_account_id    = %q
  name                = %q
  region              = "europe-west1"
  host_access_ssh_key = %q
  license             = "aos"
  aos_version         = "6.7"
  software_tier       = "pro"

  capacity   = [{ host_type = "n2-standard-32", number_of_hosts = "3" }]
  redundancy = { factor = "1" }
  network    = {
    mode              = "new"
    availability_zone = "europe-west1-b"
    project_id        = %q
  }
}
`, orgID, cloudAccountID, name, sshKey, projectID)
}
