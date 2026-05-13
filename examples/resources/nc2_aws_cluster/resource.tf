terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "organization_id" { type = string }
variable "cloud_account_id" { type = string }

resource "nc2_aws_cluster" "demo" {
  organization_id     = var.organization_id
  cloud_account_id    = var.cloud_account_id
  name                = "demo-aws"
  region              = "us-east-1"
  use_case            = "general"
  host_access_ssh_key = "demo-key"
  license             = "aos"
  aos_version         = "6.7"
  software_tier       = "pro"

  capacity = [
    {
      host_type       = "m5d.metal"
      number_of_hosts = "3"
    },
  ]

  redundancy = {
    factor = "1"
  }

  network = {
    mode              = "new"
    availability_zone = "us-east-1a"
    vpc_cidr          = "10.0.0.0/16"
  }

  resource_tags = {
    "env" = "demo"
  }
}
