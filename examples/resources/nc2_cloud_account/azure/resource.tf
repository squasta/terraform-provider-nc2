terraform {
  required_providers {
    nc2 = {
      source = "nutanix/nc2"
    }
  }
}

provider "nc2" {}

variable "organization_id" { type = string }
variable "azure_tenant_id" { type = string }
variable "azure_subscription_id" { type = string }
variable "azure_client_id" { type = string }
variable "azure_client_secret" {
  type      = string
  sensitive = true
}

resource "nc2_cloud_account" "azure_demo" {
  organization_id = var.organization_id
  cloud_provider  = "azure"
  name            = "demo-azure"
  description     = "Azure account managed by Terraform."

  credentials = {
    tenant_id       = var.azure_tenant_id
    subscription_id = var.azure_subscription_id
    client_id       = var.azure_client_id
    client_secret   = var.azure_client_secret
  }
}
