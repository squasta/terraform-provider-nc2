# Getting started

This guide walks through provisioning your first NC2 organization
and AWS cloud account from scratch.

## Prerequisites

- Terraform 1.6+ (the provider targets Plugin Framework v1.x).
- An NC2 API key, key id, and issuer (organization UUID). See the
  Nutanix portal → API Keys.
- An AWS IAM access key + secret with the policies NC2 expects
  (NC2 documentation links the canonical policy template).

## Install

```terraform
terraform {
  required_providers {
    nc2 = {
      source  = "nutanix/nc2"
      version = "~> 0.1"
    }
  }
}
```

Run `terraform init` to download the provider.

## Configure credentials

Use environment variables for the simplest setup:

```bash
export NC2_API_KEY=...
export NC2_KEY_ID=...
export NC2_ISSUER=...
```

Or use a credentials file at `~/.nc2/credentials` (INI format —
see the [authentication guide](./authentication.md)).

## First apply

```terraform
provider "nc2" {}

resource "nc2_organization" "demo" {
  name        = "demo-org"
  description = "Demo organization."
}

resource "nc2_cloud_account" "aws_demo" {
  organization_id = nc2_organization.demo.id
  cloud_provider  = "aws"
  name            = "demo-aws"

  credentials = {
    access_key_id     = var.aws_access_key_id
    secret_access_key = var.aws_secret_access_key
  }
}

resource "nc2_cloud_account_region" "us_east_1" {
  cloud_account_id = nc2_cloud_account.aws_demo.id
  region           = "us-east-1"
}
```

```bash
terraform plan
terraform apply
```

A clean plan after apply confirms drift-free behavior (FR-018).

## Next steps

- Provision your first cluster on AWS / Azure / GCP — see
  `docs/resources/nc2_aws_cluster.md` and siblings.
- Add a hibernate/resume schedule using the `desired_state`
  attribute on the `nc2_aws_cluster` resource (AWS-only per
  FR-012; not available on `nc2_azure_cluster` /
  `nc2_gcp_cluster`).
- Wire NC2 audit logs into your observability stack — see the
  [async + tasks guide](./async-and-tasks.md) for how to capture
  them.
