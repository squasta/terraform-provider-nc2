terraform {
  required_providers {
    nc2 = { source = "nutanix/nc2" }
  }
}

provider "nc2" {}

variable "task_id" { type = string }

data "nc2_task" "single" {
  id = var.task_id
}
