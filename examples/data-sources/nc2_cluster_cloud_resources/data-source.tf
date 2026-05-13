terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

variable "cluster_id" { type = string }

data "nc2_cluster_cloud_resources" "list" {
  cluster_id = var.cluster_id
}
