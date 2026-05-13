# nc2_cluster_scale_flow_gateway (Action)

Scales the cluster's Flow Gateway out (or in) to `target_node_count`.
Issues `POST /clusters/{id}/scale-out-fgw`.

## Example Usage

```terraform
action "nc2_cluster_scale_flow_gateway" "scale" {
  config {
    cluster_id        = "..."
    target_node_count = 5
  }
}
```

## Schema

### Required

- `cluster_id` (String)
- `target_node_count` (Number)
