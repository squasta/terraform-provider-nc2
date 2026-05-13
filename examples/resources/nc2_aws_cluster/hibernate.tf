resource "nc2_aws_cluster" "demo" {
  organization_id     = var.organization_id
  cloud_account_id    = var.cloud_account_id
  name                = "demo-aws"
  region              = "us-east-1"
  host_access_ssh_key = "demo-key"
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

  desired_state = "hibernated"
}
