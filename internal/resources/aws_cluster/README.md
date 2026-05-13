# `internal/resources/aws_cluster`

Implements `nc2_aws_cluster` — the AWS variant of the NC2 cluster
managed resource. Composes `internal/resources/clustershared` for
the common attributes and CRUD plumbing, adds the AWS-only
`access_policy` map, and exposes the AWS-only hibernate / resume
`desired_state` surface.

## AWS-only `access_policy` (FR-010a)

Only `nc2_aws_cluster` exposes `access_policy`. Setting the
attribute on `nc2_azure_cluster` or `nc2_gcp_cluster` is a
compile-time error because those packages do not declare it. In-place
changes route to `POST /clusters/{id}/update-access-policy` per the
clustershared router's ordering.

## AWS-only `desired_state` (FR-012)

`nc2_aws_cluster` is the only cluster resource that exposes
`desired_state` (allowed values: `running`, `hibernated`). Flipping
the value routes to `POST /clusters/{id}/hibernate` or
`POST /clusters/{id}/resume`. Combining a `desired_state` flip with
any other field change in the same `apply` is rejected at routing
time — split it into two `apply`s. The `nc2_azure_cluster` and
`nc2_gcp_cluster` resources do not expose `desired_state` at all;
this is enforced by their schemas (see their respective
`schema_test.go::TestSchema_NoDesiredState`).

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
