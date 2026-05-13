terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "cloud_account_id" { type = string }

data "nc2_cloud_account" "this" {
  id = var.cloud_account_id
}

output "cloud_account_name" {
  value = data.nc2_cloud_account.this.name
}
