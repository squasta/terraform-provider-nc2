# `internal/resources/gcp_cluster`

Implements `nc2_gcp_cluster` — the GCP variant of the NC2 cluster
managed resource. Mirrors `aws_cluster` minus the AWS-only
`access_policy` (FR-010a). GCP-specific subfields
(`network.gcp.project_id`, `network.gcp.vpc_name`) live inside the
shared `network` Map<String,String>.

## Lifecycle

- **Create** — `POST /clusters/gcp` + task polling.
- **Read** — `GET /clusters/{id}`; 404 → state removal (FR-009).
- **Update** — diff routed via `clustershared.RouteUpdate`.
- **Delete** — `POST /clusters/{id}/terminate` + polling.
- **ImportState** — `terraform import nc2_gcp_cluster.x <uuid>`.

## OpenAPI coverage

`openapi_mapping.go` claims `create_gcp`. Cross-cloud operations
are claimed by `aws_cluster` and shared transparently —
coverage-check dedups by operationId.
