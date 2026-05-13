# nc2_cloud_accounts (Data Source)

Lists every NC2 cloud account belonging to the supplied
organization. Sorted ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_cloud_accounts" "all" {
  organization_id = var.organization_id
}
```

## Schema

### Required

- `organization_id` (String) — Owning organization UUID.

### Read-only

- `cloud_accounts` (List of Object) — Each element exposes `id`,
  `name`, `cloud_provider`, `state`, `created_at`.
