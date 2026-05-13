# nc2_cloud_account_regions (Data Source)

Lists every region enabled on an NC2 cloud account. Sorted ascending
by `id` (FR-015).

## Example Usage

```terraform
data "nc2_cloud_account_regions" "all" {
  cloud_account_id = var.cloud_account_id
}
```

## Schema

### Required

- `cloud_account_id` (String) — Owning cloud account UUID.

### Read-only

- `regions` (List of Object) — Each element exposes `id`, `region`,
  `state`.
