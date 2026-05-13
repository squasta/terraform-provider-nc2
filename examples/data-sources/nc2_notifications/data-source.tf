terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

data "nc2_notifications" "all" {}
