terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

variable "cluster_id" { type = string }

action "nc2_cluster_extend_support_tunnel" "extend_8h" {
  config {
    cluster_id     = var.cluster_id
    duration_hours = 8
  }
}
