# nc2_cluster_condemn_host (Action)

Marks a specific host inside the cluster as condemned. Issues
`POST /clusters/{id}/condemn-host` and polls the returned task.

## Example Usage

```terraform
action "nc2_cluster_condemn_host" "remove_bad_host" {
  config {
    cluster_id = "00000000-0000-0000-0000-000000000000"
    host_id    = "host-abc"
  }
}
```

## Schema

### Required

- `cluster_id` (String) Target cluster UUID.
- `host_id` (String) Target host id.
