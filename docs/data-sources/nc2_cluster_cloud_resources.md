# nc2_cluster_cloud_resources (Data Source)

Lists every cloud resource NC2 created on behalf of the cluster.
Sorted ascending by `id` (FR-015).

## Example Usage

```terraform
data "nc2_cluster_cloud_resources" "list" {
  cluster_id = "00000000-0000-0000-0000-000000000000"
}
```

## Schema

### Required

- `cluster_id` (String) Parent cluster UUID.

### Read-only

- `cloud_resources` (List of Object) — `id`, `type`, `name`.
