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

resource "nc2_azure_cluster" "demo" {
  organization_id     = var.organization_id
  cloud_account_id    = var.cloud_account_id
  name                = "demo-azure"
  region              = "eastus"
  use_case            = "general"
  host_access_ssh_key = "demo-key"
  license             = "aos"
  aos_version         = "6.7"
  software_tier       = "pro"

  # NC2 on Azure only supports the AN36P and AN64 bare-metal SKUs
  # as cluster hosts. Any other Azure VM size is rejected at plan
  # time with a per-element diagnostic.
  capacity = [
    {
      host_type       = "AN36P"
      number_of_hosts = "3"
    },
  ]

  redundancy = {
    factor = "1"
  }

  network = {
    mode              = "new"
    availability_zone = "1"
    vpc_cidr          = "10.0.0.0/16"
  }
}
