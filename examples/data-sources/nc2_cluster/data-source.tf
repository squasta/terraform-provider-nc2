terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "cluster_id" { type = string }

data "nc2_cluster" "target" {
  id = var.cluster_id
}

output "cluster_state" {
  value = data.nc2_cluster.target.state
}
