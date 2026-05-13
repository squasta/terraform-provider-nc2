# `internal/resources/aws_cluster`

Implements `nc2_aws_cluster` — the AWS variant of the NC2 cluster
managed resource. Composes `internal/resources/clustershared` for
the common attributes and CRUD plumbing, adds the AWS-only
`access_policy` map.

## AWS-only `access_policy` (FR-010a)

Only `nc2_aws_cluster` exposes `access_policy`. Setting the
attribute on `nc2_azure_cluster` or `nc2_gcp_cluster` is a
compile-time error because those packages do not declare it. In-place
changes route to `POST /clusters/{id}/update-access-policy` per the
clustershared router's ordering.

## Lifecycle

- **Create** — `POST /clusters/aws` returns a task; the resource
  polls until completion via `client.PollTask`, then issues a
  follow-up `Read` to populate computed fields.
- **Read** — `GET /clusters/{id}`. A 404 removes the resource from
  state (FR-009).
- **Update** — diffed against state and routed via
  `clustershared.RouteUpdate` in this order: `update-license`,
  `update-ssh-key`, `update-capacity`, `update-resource-tags`,
  `update-access-policy`, generic `PATCH`, `hibernate`/`resume`.
- **Delete** — `POST /clusters/{id}/terminate` + `PollTask`.
- **ImportState** — `terraform import nc2_aws_cluster.x <uuid>`.

## OpenAPI coverage

`openapi_mapping.go` claims 11 operationIds spanning create, read,
all six in-place update endpoints, terminate, and hibernate/resume.

## Tests

- `resource_test.go` — sensitive registry, diff routing, OpenAPI
  coverage assertions.
- Acceptance: `tests/acceptance/nc2_aws_cluster_test.go` (gated by
  `TF_ACC=1`).
