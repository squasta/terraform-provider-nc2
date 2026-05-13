terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "organization_id" { type = string }

data "nc2_cloud_accounts" "all" {
  organization_id = var.organization_id
}

output "cloud_account_ids" {
  value = [for a in data.nc2_cloud_accounts.all.cloud_accounts : a.id]
}
