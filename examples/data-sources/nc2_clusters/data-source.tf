terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

data "nc2_clusters" "all" {}

output "cluster_names" {
  value = [for c in data.nc2_clusters.all.clusters : c.name]
}
