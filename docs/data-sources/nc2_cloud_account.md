# nc2_cloud_account (Data Source)

Fetches a single NC2 cloud account by id. Credentials are NOT
returned.

## Example Usage

```terraform
data "nc2_cloud_account" "this" {
  id = var.cloud_account_id
}
```

## Schema

### Required

- `id` (String) — Cloud account UUID.

### Read-only

- `organization_id`, `cloud_provider`, `name`, `description`,
  `state`, `created_at`, `updated_at`.
