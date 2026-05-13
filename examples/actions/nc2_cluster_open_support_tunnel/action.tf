terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

variable "cluster_id" { type = string }

action "nc2_cluster_open_support_tunnel" "open_for_support" {
  config {
    cluster_id = var.cluster_id
  }
}
