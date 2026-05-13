# nc2_remote_storage_profiles (Data Source)

Lists remote-storage profiles inside a cloud-account region. Sorted
ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_remote_storage_profiles" "list" {
  cloud_account_id = "00000000-0000-0000-0000-000000000000"
  region_id        = "us-east-1"
}
```

## Schema

### Required

- `cloud_account_id` (String)
- `region_id` (String)

### Read-only

- `remote_storage_profiles` (List of Object) — `id`, `name`, `capacity`.
