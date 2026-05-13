package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_AzureCluster_Lifecycle covers FR-023 for nc2_azure_cluster.
func TestAcc_AzureCluster_Lifecycle(t *testing.T) {
	SkipIfNotAcc(t)
	envs := SkipIfMissing(t,
		"NC2_API_KEY", "NC2_KEY_ID", "NC2_ISSUER",
		"NC2_TEST_ORGANIZATION_ID",
		"NC2_TEST_AZURE_CLOUD_ACCOUNT_ID",
		"NC2_TEST_AZURE_SSH_KEY_NAME",
	)
	orgID, cloudAccountID, sshKey := envs[3], envs[4], envs[5]
	name := fmt.Sprintf("tf-acc-azure-%d", randInt())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: azureClusterConfig(orgID, cloudAccountID, name, sshKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nc2_azure_cluster.target", "name", name),
				),
			},
		},
	})
}

func azureClusterConfig(orgID, cloudAccountID, name, sshKey string) string {
	return fmt.Sprintf(`
resource "nc2_azure_cluster" "target" {
  organization_id     = %q
  cloud_account_id    = %q
  name                = %q
  region              = "eastus"
  host_access_ssh_key = %q
  license             = "aos"
  aos_version         = "6.7"
  software_tier       = "pro"

  capacity   = [{ host_type = "AN36P", number_of_hosts = "3" }]
  redundancy = { factor = "1" }
  network    = {
    mode              = "new"
    availability_zone = "1"
    vpc_cidr          = "10.0.0.0/16"
  }
}
`, orgID, cloudAccountID, name, sshKey)
}
