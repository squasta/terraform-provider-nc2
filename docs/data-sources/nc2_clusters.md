# nc2_clusters (Data Source)

Lists every NC2 cluster visible to the configured credentials.
Sorted ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_clusters" "all" {}
```

## Schema

### Read-only

- `clusters` (List of Object) — each entry exposes `id`, `name`,
  `organization_id`, `cloud_account_id`, `cloud_provider`, `region`,
  `state`, `created_at`.
