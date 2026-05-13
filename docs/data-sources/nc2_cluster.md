# nc2_cluster (Data Source)

Looks up a single NC2 cluster by `id`.

## Example Usage

```terraform
data "nc2_cluster" "target" {
  id = "00000000-0000-0000-0000-000000000000"
}
```

## Schema

### Required

- `id` (String) Cluster UUID.

### Read-only

- `name`, `organization_id`, `cloud_account_id`, `cloud_provider`,
  `region`, `state`, `created_at`.
