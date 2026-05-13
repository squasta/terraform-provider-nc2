package acceptance

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAcc_AWSCluster_HibernateResume covers FR-023 / US4 for the
// AWS variant: running → hibernated → running with plan-clean
// after each apply.
func TestAcc_AWSCluster_HibernateResume(t *testing.T) {
	SkipIfNotAcc(t)
	envs := SkipIfMissing(t,
		"NC2_API_KEY", "NC2_KEY_ID", "NC2_ISSUER",
		"NC2_TEST_ORGANIZATION_ID",
		"NC2_TEST_AWS_CLOUD_ACCOUNT_ID",
		"NC2_TEST_AWS_SSH_KEY_NAME",
	)
	orgID, cloudAccountID, sshKey := envs[3], envs[4], envs[5]
	name := fmt.Sprintf("tf-acc-aws-h-%d", randInt())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: awsClusterConfigDesired(orgID, cloudAccountID, name, sshKey, "running")},
			{Config: awsClusterConfigDesired(orgID, cloudAccountID, name, sshKey, "hibernated")},
			{Config: awsClusterConfigDesired(orgID, cloudAccountID, name, sshKey, "running")},
		},
	})
}

func awsClusterConfigDesired(orgID, cloudAccountID, name, sshKey, desired string) string {
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

  capacity   = [{ host_type = "m5d.metal", number_of_hosts = "3" }]
  redundancy = { factor = "1" }
  network    = {
    mode              = "new"
    availability_zone = "us-east-1a"
    vpc_cidr          = "10.0.0.0/16"
  }

  desired_state = %q
}
`, orgID, cloudAccountID, name, sshKey, desired)
}
