# `internal/resources/gcp_cluster`

Implements `nc2_gcp_cluster` — the GCP variant of the NC2 cluster
managed resource. Mirrors `aws_cluster` minus the AWS-only
`access_policy` (FR-010a) **and** minus the AWS-only `desired_state`
hibernate / resume surface (FR-012). GCP-specific subfields
(`network.gcp.project_id`, `network.gcp.vpc_name`) live inside the
shared `network` Map<String,String>.

## Hibernate / resume

Hibernate and resume are **not** supported on `nc2_gcp_cluster`.
The `desired_state` attribute is intentionally absent; setting it in
configuration produces a schema error at plan time.

## Lifecycle

- **Create** — `POST /clusters/gcp` + task polling.
- **Read** — `GET /clusters/{id}`; 404 → state removal (FR-009).
- **Update** — diff routed via `clustershared.RouteUpdate`. The
  router never schedules `hibernate` / `resume` because
  `computeDiff` leaves `Diff.DesiredStateFrom/To` empty.
- **Delete** — `POST /clusters/{id}/terminate` + polling.
- **ImportState** — `terraform import nc2_gcp_cluster.x <uuid>`.

## OpenAPI coverage

`openapi_mapping.go` claims `create_gcp`. Cross-cloud operations
are claimed by `aws_cluster` and shared transparently —
coverage-check dedups by operationId.
