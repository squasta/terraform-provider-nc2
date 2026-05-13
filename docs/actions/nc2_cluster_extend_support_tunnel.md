# nc2_cluster_extend_support_tunnel (Action)

Extends an open Support Tunnel by `duration_hours`. Issues
`POST /clusters/{id}/extend-support-tunnel`.

## Example Usage

```terraform
action "nc2_cluster_extend_support_tunnel" "extend" {
  config {
    cluster_id     = "..."
    duration_hours = 8
  }
}
```

## Schema

### Required

- `cluster_id` (String)
- `duration_hours` (Number)
