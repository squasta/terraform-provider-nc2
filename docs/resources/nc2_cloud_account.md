# nc2_cloud_account (Resource)

Manages a credentialed AWS / Azure / GCP cloud account that NC2
uses to provision infrastructure.

## ⚠️ Destroy semantics (FR-006)

NC2 has **no DELETE endpoint** for cloud accounts. `terraform destroy`
removes the resource from state only; the underlying NC2 cloud
account remains live until removed via the NC2 UI. A plan-time
warning is emitted when destroy is planned.

## Example Usage

```terraform
resource "nc2_cloud_account" "aws_demo" {
  organization_id = nc2_organization.demo.id
  cloud_provider  = "aws"
  name            = "demo-aws"

  credentials = {
    access_key_id     = var.aws_access_key_id
    secret_access_key = var.aws_secret_access_key
  }
}
```

## Schema

### Required

- `organization_id` (String) — Owning organization UUID.
  RequiresReplace.
- `cloud_provider` (String) — `aws` | `azure` | `gcp`.
  RequiresReplace.
- `name` (String) — 1–128 characters.
- `credentials` (Map of String, Sensitive) — Per-cloud credential
  shape. The full subtree is sensitive and never written to logs or
  plan output.

### Optional

- `description` (String) — ≤ 2048 characters.

### Read-only

- `id`, `state`, `created_at`, `updated_at`.

## Update routing

- `credentials` change → `POST /cloud-accounts/{id}/update-credentials`.
- `name` / `description` change → `PATCH /cloud-accounts/{id}`.
- Both → credentials first, then PATCH (deterministic order).

## Import

```shell
terraform import nc2_cloud_account.aws_demo <cloud-account-uuid>
```

After import, write your `credentials` block — NC2 will not echo
existing credential values; the next plan will show that field as
needing reconfiguration.
