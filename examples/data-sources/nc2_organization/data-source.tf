terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "organization_id" { type = string }

data "nc2_organization" "this" {
  id = var.organization_id
}

output "organization_name" {
  value = data.nc2_organization.this.name
}
