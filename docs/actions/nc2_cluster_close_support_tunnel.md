# nc2_cluster_close_support_tunnel (Action)

Closes the open Support Tunnel on the cluster. Issues
`POST /clusters/{id}/close-support-tunnel`.

## Example Usage

```terraform
action "nc2_cluster_close_support_tunnel" "close" {
  config { cluster_id = "..." }
}
```

## Schema

### Required

- `cluster_id` (String) Target cluster UUID.
