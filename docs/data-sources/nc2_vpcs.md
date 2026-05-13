# nc2_vpcs (Data Source)

Lists AWS VPCs inside a cloud-account region. Sorted ascending by
`id` (FR-015).

## Example Usage

```terraform
data "nc2_vpcs" "list" {
  cloud_account_id = "00000000-0000-0000-0000-000000000000"
  region_id        = "us-east-1"
}
```

## Schema

### Required

- `cloud_account_id` (String)
- `region_id` (String)

### Read-only

- `vpcs` (List of Object) — `id`, `name`, `cidr`.
