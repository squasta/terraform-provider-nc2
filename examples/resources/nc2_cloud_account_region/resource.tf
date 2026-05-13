terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "cloud_account_id" { type = string }
variable "region" {
  type    = string
  default = "us-east-1"
}

resource "nc2_cloud_account_region" "demo" {
  cloud_account_id = var.cloud_account_id
  region           = var.region
}
