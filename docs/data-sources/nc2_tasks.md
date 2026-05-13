# nc2_tasks (Data Source)

Lists every NC2 task visible to the configured credentials. Sorted
ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_tasks" "all" {}
```

## Schema

### Read-only

- `tasks` (List of Object) — `id`, `status`, `cluster_id`,
  `created_at`, `ended_at`.
