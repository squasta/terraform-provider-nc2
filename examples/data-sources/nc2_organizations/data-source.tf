terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

data "nc2_organizations" "all" {}

output "organization_ids" {
  value = [for o in data.nc2_organizations.all.organizations : o.id]
}
