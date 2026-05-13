# nc2_prism_centrals (Data Source)

Lists Prism Central instances inside a cloud-account region. Sorted
ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_prism_centrals" "list" {
  cloud_account_id = "00000000-0000-0000-0000-000000000000"
  region_id        = "us-east-1"
}
```

## Schema

### Required

- `cloud_account_id` (String)
- `region_id` (String)

### Read-only

- `prism_centrals` (List of Object) — `id`, `name`, `version`, `state`.
