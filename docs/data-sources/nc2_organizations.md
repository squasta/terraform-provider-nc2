# nc2_organizations (Data Source)

Lists every NC2 organization visible to the configured credentials.
Sorted ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_organizations" "all" {}

output "organization_ids" {
  value = [for o in data.nc2_organizations.all.organizations : o.id]
}
```

## Schema

### Read-only

- `organizations` (List of Object) — Each element exposes `id`,
  `name`, `description`, `state`, `created_at`.
