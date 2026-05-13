terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

variable "cluster_id" { type = string }

action "nc2_cluster_scale_flow_gateway" "scale_to_5" {
  config {
    cluster_id        = var.cluster_id
    target_node_count = 5
  }
}
