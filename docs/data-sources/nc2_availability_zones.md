# nc2_availability_zones (Data Source)

Lists availability zones inside a cloud-account region. Sorted
ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_availability_zones" "list" {
  cloud_account_id = "00000000-0000-0000-0000-000000000000"
  region_id        = "us-east-1"
}
```

## Schema

### Required

- `cloud_account_id` (String) Parent cloud account UUID.
- `region_id` (String) Parent region id.

### Read-only

- `availability_zones` (List of Object) — `id`, `name`, `state`.
