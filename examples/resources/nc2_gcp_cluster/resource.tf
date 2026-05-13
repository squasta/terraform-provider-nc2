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
variable "gcp_project_id" { type = string }

resource "nc2_gcp_cluster" "demo" {
  organization_id     = var.organization_id
  cloud_account_id    = var.cloud_account_id
  name                = "demo-gcp"
  region              = "europe-west1"
  use_case            = "general"
  host_access_ssh_key = "demo-key"
  license             = "aos"
  aos_version         = "6.7"
  software_tier       = "pro"

  capacity = [
    {
      host_type       = "n2-standard-32"
      number_of_hosts = "3"
    },
  ]

  redundancy = {
    factor = "1"
  }

  network = {
    mode              = "new"
    availability_zone = "europe-west1-b"
    project_id        = var.gcp_project_id
  }
}
