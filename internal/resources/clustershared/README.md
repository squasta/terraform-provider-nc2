# `internal/resources/clustershared`

Shared, **pure** helpers for the three cloud-specific cluster resource
packages — `aws_cluster`, `azure_cluster`, `gcp_cluster`. Every
function is I/O free so it can be unit-tested without a fake server
and trivially reasoned about in the framework's plan/apply loop.

## Public API

- `type OperationKind string` and the named constants
  `OpUpdateLicense`, `OpUpdateSSHKey`, `OpUpdateCapacity`,
  `OpUpdateResourceTags`, `OpUpdateAccessPolicy`, `OpGenericPatch`,
  `OpHibernate`, `OpResume`.
- `type Operation struct{ Kind OperationKind }` — what RouteUpdate
  schedules.
- `type Diff struct{ ... }` — boolean per attribute group + the
  before/after of `desired_state`.
- `func RouteUpdate(d Diff) ([]Operation, error)` — the deterministic
  router.
- `func CommonAttributeNames() []string` — sorted reference list of
  every attribute that the per-cloud schemas share.

## Update routing precedence (data-model.md §4)

When multiple changes appear in the same diff, RouteUpdate emits
operations in this fixed order to maximize the chance of a successful
end-to-end apply:

1. `update-license`
2. `update-ssh-key`
3. `update-capacity`
4. `update-resource-tags`
5. `update-access-policy` (AWS only — non-AWS callers leave
   `Diff.AccessPolicy = false`)
6. generic `PATCH /clusters/{id}`
7. `hibernate` / `resume` if `desired_state` changed

## Plan-time guard rails

`RouteUpdate` returns an error if either:

- `desired_state` changes alongside any other field in the same
  diff (the user must `terraform apply -target` in two steps), or
- `desired_state` is set to a value other than `running` or
  `hibernated`.

## Tests

`helpers_test.go` covers ordering, no-op, hibernate-only,
resume-only, combined-diff rejection, and bad-value rejection. The
schema-shape test in each per-cloud `aws_cluster/`, `azure_cluster/`,
`gcp_cluster/` package consumes `CommonAttributeNames()` to enforce
the shared schema.
