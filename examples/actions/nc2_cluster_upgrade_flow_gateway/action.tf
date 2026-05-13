terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

variable "cluster_id" { type = string }

action "nc2_cluster_upgrade_flow_gateway" "upgrade_fgw" {
  config {
    cluster_id = var.cluster_id
  }
}
