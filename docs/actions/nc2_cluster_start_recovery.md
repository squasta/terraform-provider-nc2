# nc2_cluster_start_recovery (Action)

Starts a recovery operation for the cluster. Issues
`POST /clusters/{id}/start-recovery`.

## Example Usage

```terraform
action "nc2_cluster_start_recovery" "recover" {
  config { cluster_id = "..." }
}
```

## Schema

### Required

- `cluster_id` (String) Target cluster UUID.
