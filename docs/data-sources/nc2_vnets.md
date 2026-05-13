# nc2_vnets (Data Source)

Lists Azure / GCP virtual networks inside a cloud-account region.
Sorted ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_vnets" "list" {
  cloud_account_id = "00000000-0000-0000-0000-000000000000"
  region_id        = "eastus"
}
```

## Schema

### Required

- `cloud_account_id` (String)
- `region_id` (String)

### Read-only

- `vnets` (List of Object) — `id`, `name`, `cidr`.
