# nc2_cluster_open_support_tunnel (Action)

Opens a Support Tunnel on the cluster. Issues
`POST /clusters/{id}/open-support-tunnel` and polls the returned task.

## Example Usage

```terraform
action "nc2_cluster_open_support_tunnel" "open" {
  config { cluster_id = "..." }
}
```

## Schema

### Required

- `cluster_id` (String) Target cluster UUID.
