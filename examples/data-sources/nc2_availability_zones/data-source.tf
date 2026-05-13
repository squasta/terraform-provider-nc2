terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

variable "cloud_account_id" { type = string }
variable "region_id" { type = string }

data "nc2_availability_zones" "list" {
  cloud_account_id = var.cloud_account_id
  region_id        = var.region_id
}

output "az_names" {
  value = [for az in data.nc2_availability_zones.list.availability_zones : az.name]
}
