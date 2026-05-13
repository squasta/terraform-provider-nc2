# nc2_organization_audit_trail (Data Source)

Lists audit-trail entries recorded for an NC2 organization. Sorted
ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_organization_audit_trail" "recent" {
  organization_id = var.organization_id
}
```

## Schema

### Required

- `organization_id` (String) — Owning organization UUID.

### Read-only

- `entries` (List of Object) — Each element exposes `id`, `actor`,
  `action`, `target`, `created_at`.
