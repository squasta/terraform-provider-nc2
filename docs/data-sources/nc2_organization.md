# nc2_organization (Data Source)

Fetches a single NC2 organization by id.

## Example Usage

```terraform
data "nc2_organization" "this" {
  id = var.organization_id
}
```

## Schema

### Required

- `id` (String) — Organization UUID.

### Read-only

- `name`, `description`, `state`, `created_at`, `updated_at`.
