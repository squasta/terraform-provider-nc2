# nc2_cloud_account_region (Resource)

Enables a cloud-provider region on an NC2 cloud account.

## ⚠️ Destroy semantics (FR-007)

NC2 has **no DELETE endpoint** for cloud account regions.
`terraform destroy` removes the resource from state only; the
region remains enabled in NC2 until you disable it via the NC2 UI.

## Example Usage

```terraform
resource "nc2_cloud_account_region" "demo" {
  cloud_account_id = nc2_cloud_account.aws_demo.id
  region           = "us-east-1"
}
```

## Schema

### Required

- `cloud_account_id` (String) — Owning cloud account UUID.
  RequiresReplace.
- `region` (String) — Cloud-provider region identifier
  (`us-east-1`, `eastus`, `europe-west1`, etc.). RequiresReplace.

### Read-only

- `id` (String) — Region resource id.
- `state` (String) — `available` | `disabled`.

## Import

```shell
terraform import nc2_cloud_account_region.demo <cloud-account-uuid>/<region-id>
```

The composite import id is required because NC2 has no
`/regions/{id}` endpoint; the cloud account context is needed to
look the region up.
