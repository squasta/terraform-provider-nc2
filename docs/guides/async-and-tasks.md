# Asynchronous operations & tasks

Most NC2 mutating operations (cluster create/update/delete, cloud
account create, action invocations) are **asynchronous**: the API
returns a `task_id` and Terraform must poll
`GET /tasks/{id}` until the task reaches a terminal state.

This guide explains the model, the configurable knobs, and how to
troubleshoot.

## Task lifecycle

```
pending → running → done       (success)
                  → failed     (terminal error)
                  → cancelled  (operator/system cancel)
```

The provider's `internal/client.PollTask` waits for any of `done`,
`failed`, or `cancelled`. On `failed` or `cancelled` it returns a
typed `APIError` carrying the NC2 error code and human-readable
message.

## Configuration

```terraform
provider "nc2" {
  task_poll_interval_seconds = 5    # default
  task_max_timeout_seconds   = 3600 # default
}
```

- `task_poll_interval_seconds`: how often the provider polls the
  task endpoint while the task is non-terminal.
- `task_max_timeout_seconds`: cap on total wait time per task.
  When exceeded, the provider returns a timeout error and includes
  the `task_id` so you can inspect it manually via the NC2 UI or
  the `nc2_task` data source.

## Reading task state

Outside the lifecycle of a managed resource you can read tasks
directly:

```terraform
data "nc2_task" "x" {
  id = "task-uuid"
}

output "status" {
  value = data.nc2_task.x.status
}
```

`data.nc2_tasks` returns a sorted-by-id list of every task visible
to the configured credentials.

## Audit records

Every async call emits one audit record at request time and a
follow-up at task completion:

- `terraform_op = nc2_aws_cluster.Create`,
  `http_status = 202`, `nc2_task_id = ...`.
- `terraform_op = nc2_aws_cluster.Create`,
  `nc2_task_id = ...`, status from the final task body.

Filter your log store by `correlation_id` to walk a specific call
end-to-end.

## Troubleshooting

| Symptom | Cause | Action |
|---|---|---|
| Cluster create hangs past `task_max_timeout_seconds` | NC2 is genuinely still working OR a long-tail provisioning step | Read `data.nc2_task.<id>` for `status`/`progress`; the cluster id is on the task body once Provisioning starts. |
| `task failed (NC2 error code)` | Cloud-side rejection (quota, IAM, network) | The NC2 error message is verbatim from the API; cross-reference with the cloud provider's console. |
| Repeated 401 retries before each call | Clock skew between the operator and NC2 (JWT exp window is 5 min) | Sync the operator clock; the provider already retries once on 401 with a forced JWT refresh. |
| 404 during Read after Create succeeds | Eventual consistency window between control-plane writes and reads | Re-run the plan; the provider transparently treats a 404 on Read as state-removal once the resource is genuinely gone (FR-009). |

## Hibernate / resume (US4, AWS only)

Setting `desired_state = "hibernated"` (or `"running"`) on the
`nc2_aws_cluster` resource routes to the dedicated hibernate /
resume endpoints, which are themselves async. The provider
intentionally **rejects** combining a `desired_state` flip with
other field changes in the same `apply` (split into two `apply`s
instead); this avoids the failure mode where one-of-N
sub-operations succeeds and leaves the cluster in an indeterminate
state.

> Hibernate / resume is AWS-only (FR-012). The `nc2_azure_cluster`
> and `nc2_gcp_cluster` resources do not expose `desired_state`;
> referencing it in configuration is a schema error at plan time.
