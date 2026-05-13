# nc2_notifications (Data Source)

Lists every NC2 notification visible to the configured credentials.
Sorted ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_notifications" "all" {}
```

## Schema

### Read-only

- `notifications` (List of Object) — `id`, `severity`, `message`,
  `acknowledged`, `created_at`.
