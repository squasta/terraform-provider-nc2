terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "cloud_account_id" { type = string }

data "nc2_cloud_account_regions" "all" {
  cloud_account_id = var.cloud_account_id
}

output "region_ids" {
  value = [for r in data.nc2_cloud_account_regions.all.regions : r.id]
}
