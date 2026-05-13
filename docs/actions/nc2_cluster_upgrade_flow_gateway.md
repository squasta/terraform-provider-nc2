# nc2_cluster_upgrade_flow_gateway (Action)

Upgrades the cluster's Flow Gateway. Issues
`POST /clusters/{id}/upgrade-fgw`.

## Example Usage

```terraform
action "nc2_cluster_upgrade_flow_gateway" "upgrade" {
  config { cluster_id = "..." }
}
```

## Schema

### Required

- `cluster_id` (String) Target cluster UUID.
