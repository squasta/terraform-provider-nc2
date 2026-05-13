terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

variable "notification_id" { type = string }

action "nc2_notification_acknowledge" "ack" {
  config {
    notification_id = var.notification_id
  }
}
