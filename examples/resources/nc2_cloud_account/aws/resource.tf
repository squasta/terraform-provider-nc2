terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "organization_id" { type = string }
variable "aws_access_key_id" {
  type      = string
  sensitive = true
}
variable "aws_secret_access_key" {
  type      = string
  sensitive = true
}
variable "aws_role_arn" {
  type    = string
  default = ""
}

resource "nc2_cloud_account" "aws_demo" {
  organization_id = var.organization_id
  cloud_provider  = "aws"
  name            = "demo-aws"
  description     = "AWS account managed by Terraform."

  credentials = {
    access_key_id     = var.aws_access_key_id
    secret_access_key = var.aws_secret_access_key
    role_arn          = var.aws_role_arn
  }
}

output "cloud_account_id" {
  value = nc2_cloud_account.aws_demo.id
}
