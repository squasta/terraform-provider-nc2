# nc2_task (Data Source)

Looks up a single NC2 task by `id`.

## Example Usage

```terraform
data "nc2_task" "single" {
  id = "00000000-0000-0000-0000-000000000000"
}
```

## Schema

### Required

- `id` (String) Task id.

### Read-only

- `status`, `cluster_id`, `created_at`, `ended_at`,
  `error_code`, `error_message`.
