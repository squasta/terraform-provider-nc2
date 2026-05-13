# `internal/resources/clustershared`

Shared, **pure** helpers for the three cloud-specific cluster resource
packages — `aws_cluster`, `azure_cluster`, `gcp_cluster`. Every
function is I/O free so it can be unit-tested without a fake server
and trivially reasoned about in the framework's plan/apply loop.

## Public API

- `type Model` — the cross-cloud framework model. Embedded by
  `azure_cluster.model` and `gcp_cluster.model`.
- `type HibernatingModel` — `Model` extended with the AWS-only
  `desired_state` attribute. Embedded by `aws_cluster.model`
  (FR-012).
- `type OperationKind string` and the named constants
  `OpUpdateLicense`, `OpUpdateSSHKey`, `OpUpdateCapacity`,
  `OpUpdateResourceTags`, `OpUpdateAccessPolicy`, `OpGenericPatch`,
  `OpHibernate`, `OpResume`.
- `type Operation struct{ Kind OperationKind }` — what RouteUpdate
  schedules.
- `type Diff struct{ ... }` — boolean per attribute group + the
  before/after of `desired_state` (only AWS populates the latter).
- `func RouteUpdate(d Diff) ([]Operation, error)` — the deterministic
  router.
- `func CommonAttributes() map[string]schema.Attribute` — shared
  schema baseline (no `desired_state`).
- `func CommonAttributesWithHibernate() map[string]schema.Attribute`
  — adds the AWS-only `desired_state` attribute (FR-012).
- `func CommonAttributeNames() []string` — sorted reference list of
  every attribute that the per-cloud schemas share.
- `func DeriveDesiredStateIfUnset(observedState string, current types.String) types.String`
  — pure helper used by `nc2_aws_cluster` to map the observed
  cluster state back into the Optional+Computed `desired_state`
  attribute after each Read.
- `func ReadHibernatingCluster` — AWS convenience wrapper around
  `ReadCluster` that runs `DeriveDesiredStateIfUnset` afterwards.
- `const MinClusterHosts = 1`, `MinProductionClusterHosts = 3`,
  `MaxClusterHosts = 28` — the NC2 cluster sizing rule.
- `func AllowedClusterHostCounts() []int`,
  `func IsAllowedClusterHostCount(int) bool` — pure helpers
  exposing the allowed totals (1, then 3..28 inclusive — never 2).
- `func ValidateCapacityHostCount(types.List) diag.Diagnostics`
  — plan-time validator called from each cluster resource's
  `ValidateConfig`. Emits per-element diagnostics for parse
  failures / `< 1` values, plus an aggregate diagnostic when the
  sum of all known `number_of_hosts` violates the rule. Skips the
  aggregate when any element is unknown so it never produces a
  false positive on partially-computed plans.

## Update routing precedence (data-model.md §4)

When multiple changes appear in the same diff, RouteUpdate emits
operations in this fixed order to maximize the chance of a successful
end-to-end apply:

1. `update-license`
2. `update-ssh-key`
3. `update-capacity`
4. `update-resource-tags`
5. `update-access-policy` (AWS only, FR-010a — non-AWS callers
   leave `Diff.AccessPolicy = false`)
6. generic `PATCH /clusters/{id}`
7. `hibernate` / `resume` if `desired_state` changed (AWS only,
   FR-012 — non-AWS callers leave `Diff.DesiredStateFrom/To`
   empty, so this branch is unreachable from Azure / GCP)

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
