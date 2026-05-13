terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "organization_id" { type = string }

data "nc2_organization_audit_trail" "recent" {
  organization_id = var.organization_id
}

output "audit_entries" {
  value = data.nc2_organization_audit_trail.recent.entries
}
