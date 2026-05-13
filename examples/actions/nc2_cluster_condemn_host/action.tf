terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

variable "cluster_id" { type = string }
variable "host_id" { type = string }

action "nc2_cluster_condemn_host" "remove_bad_host" {
  config {
    cluster_id = var.cluster_id
    host_id    = var.host_id
  }
}
