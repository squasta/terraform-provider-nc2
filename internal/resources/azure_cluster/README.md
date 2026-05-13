# `internal/resources/azure_cluster`

Implements `nc2_azure_cluster` — the Azure variant of the NC2
cluster managed resource. Mirrors `aws_cluster` minus the AWS-only
`access_policy` (FR-010a) **and** minus the AWS-only `desired_state`
hibernate / resume surface (FR-012). Azure-specific subnet / vnet
attributes live inside the shared `network` Map<String,String>.

## Hibernate / resume

Hibernate and resume are **not** supported on `nc2_azure_cluster`.
The `desired_state` attribute is intentionally absent; setting it in
configuration produces a schema error at plan time. If you need to
shut down an Azure cluster temporarily, terminate and recreate it
through Terraform (NC2 does not currently support hibernate on
Azure-hosted clusters in a way the provider can model
deterministically — FR-012).

## Lifecycle

- **Create** — `POST /clusters/azure` + task polling.
- **Read** — `GET /clusters/{id}`; 404 → state removal (FR-009).
- **Update** — diff routed via `clustershared.RouteUpdate`. The
  router never schedules `hibernate` / `resume` because
  `computeDiff` leaves `Diff.DesiredStateFrom/To` empty.
- **Delete** — `POST /clusters/{id}/terminate` + polling.
- **ImportState** — `terraform import nc2_azure_cluster.x <uuid>`.

## OpenAPI coverage

`openapi_mapping.go` claims `create_azure`. The cross-cloud
operations (`show`, `update`, `update_capacity`, `update_ssh_key`,
`update_license`, `update_resource_tags`, `terminate`, `hibernate`,
`resume`) are claimed by `aws_cluster` and shared transparently —
coverage-check dedups by operationId.

## Tests

- `resource_test.go` — covers diff routing, AccessPolicy absence,
  OpenAPI mapping.
- Acceptance: `tests/acceptance/nc2_azure_cluster_test.go`.
