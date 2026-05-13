package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_AWSCluster_Lifecycle covers FR-023 + FR-018 + FR-019 +
// US2 in-place-update scenarios for nc2_aws_cluster.
func TestAcc_AWSCluster_Lifecycle(t *testing.T) {
	SkipIfNotAcc(t)
	envs := SkipIfMissing(t,
		"NC2_API_KEY", "NC2_KEY_ID", "NC2_ISSUER",
		"NC2_TEST_ORGANIZATION_ID",
		"NC2_TEST_AWS_CLOUD_ACCOUNT_ID",
		"NC2_TEST_AWS_SSH_KEY_NAME",
	)
	orgID, cloudAccountID, sshKey := envs[3], envs[4], envs[5]
	name := fmt.Sprintf("tf-acc-aws-%d", randInt())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: awsClusterConfig(orgID, cloudAccountID, name, sshKey, "3"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("nc2_aws_cluster.target", "name", name),
				),
			},
			{
				Config: awsClusterConfig(orgID, cloudAccountID, name, sshKey, "5"),
			},
		},
	})
}

func awsClusterConfig(orgID, cloudAccountID, name, sshKey, hostCount string) string {
	return fmt.Sprintf(`
resource "nc2_aws_cluster" "target" {
  organization_id     = %q
  cloud_account_id    = %q
  name                = %q
  region              = "us-east-1"
  host_access_ssh_key = %q
  license             = "aos"
  aos_version         = "6.7"
  software_tier       = "pro"

  capacity   = [{ host_type = "m5d.metal", number_of_hosts = %q }]
  redundancy = { factor = "1" }
  network    = {
    mode              = "new"
    availability_zone = "us-east-1a"
    vpc_cidr          = "10.0.0.0/16"
  }
}
`, orgID, cloudAccountID, name, sshKey, hostCount)
}
