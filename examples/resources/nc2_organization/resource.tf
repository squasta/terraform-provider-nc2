terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

resource "nc2_organization" "demo" {
  name        = "demo-org"
  description = "Demo organization managed by Terraform."
}

output "organization_id" {
  value = nc2_organization.demo.id
}
