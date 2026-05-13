terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "organization_id" { type = string }
variable "gcp_service_account_json" {
  type      = string
  sensitive = true
}

resource "nc2_cloud_account" "gcp_demo" {
  organization_id = var.organization_id
  cloud_provider  = "gcp"
  name            = "demo-gcp"
  description     = "GCP account managed by Terraform."

  credentials = {
    service_account_json = var.gcp_service_account_json
  }
}
