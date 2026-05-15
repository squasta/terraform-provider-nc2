# Changelog

All notable changes to the `terraform-provider-nc2` provider will be
documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **BREAKING (pre-1.0)**: Minimum Go version required to build the
  provider is now **Go 1.26.3** (raised from `go 1.25.8`). The
  `go.mod` `go` directive now matches the existing `toolchain
  go1.26.3` directive, so consumers building from source with an
  older Go release will see the standard `requires go >= 1.26.3`
  error from the toolchain. Rationale: align the declared minimum
  with the toolchain we actually use; remove stdlib CVEs that the
  release-pipeline `osv-scanner` gate was flagging against the
  older runtime; reduce the chance of subtle 1.26-only API usage
  silently relying on the toolchain auto-download. CI's matrix has
  been narrowed to `["1.26.x"]` to match.
- **BREAKING (pre-1.0)**: Hibernate / resume (`desired_state`) is
  now AWS-only on `nc2_aws_cluster` (FR-012). The attribute has
  been removed from `nc2_azure_cluster` and `nc2_gcp_cluster`;
  setting `desired_state` on those resources now produces a schema
  error at plan time. The shared `clustershared.Model` no longer
  carries the `DesiredState` field — AWS embeds the new
  `clustershared.HibernatingModel` instead, which adds the
  attribute back.
- **BREAKING (pre-1.0)**: `nc2_azure_cluster` now rejects any
  `capacity[*].host_type` outside the supported NC2-on-Azure
  bare-metal SKUs (`AN36P`, `AN64`) at plan time via
  `ValidateConfig`. Existing configurations referencing generic
  Azure VM sizes (e.g. `Standard_D32s_v4`) or AWS-style SKUs must
  be updated; previous releases would have surfaced the failure
  only at apply time as an opaque NC2 API error.
- All three cluster resources (`nc2_aws_cluster`,
  `nc2_azure_cluster`, `nc2_gcp_cluster`) now reject unsupported
  cluster sizes at plan time. The sum of
  `capacity[*].number_of_hosts` must equal `1` or fall in `3..28`
  inclusive — a 2-host cluster (no quorum) and totals `0` or
  `> 28` are rejected with an attribute-rooted diagnostic. The
  per-element / aggregate validator lives in
  `internal/resources/clustershared/ValidateCapacityHostCount` so
  the three resources cannot drift.

### Added

- US1 (`nc2_organization`, `nc2_cloud_account`,
  `nc2_cloud_account_region`) managed resources and the six
  read-side data sources (`nc2_organizations`, `nc2_organization`,
  `nc2_organization_audit_trail`, `nc2_cloud_accounts`,
  `nc2_cloud_account`, `nc2_cloud_account_regions`).
- US2 (`nc2_aws_cluster`, `nc2_azure_cluster`, `nc2_gcp_cluster`)
  cluster managed resources with full update routing per FR-010
  and the cluster data sources (`nc2_clusters`, `nc2_cluster`,
  `nc2_cluster_cloud_resources`).
- US3 inventory data sources (`nc2_availability_zones`,
  `nc2_vpcs`, `nc2_vnets`, `nc2_ssh_keys`, `nc2_prism_centrals`,
  `nc2_remote_storage_profiles`, `nc2_notifications`, `nc2_tasks`,
  `nc2_task`).
- US4 hibernate/resume via the `desired_state` attribute on every
  cluster resource (FR-011).
- US5 cluster-scoped operational actions (`nc2_cluster_condemn_host`,
  `nc2_cluster_open_support_tunnel`,
  `nc2_cluster_extend_support_tunnel`,
  `nc2_cluster_close_support_tunnel`,
  `nc2_cluster_scale_flow_gateway`,
  `nc2_cluster_upgrade_flow_gateway`,
  `nc2_cluster_start_recovery`) and `nc2_notification_acknowledge`.
- Strict TLS verification with optional `ca_bundle` (FR-003b/c).
- Sensitive-field redaction in plan output, state, and audit logs
  (FR-002 + per-resource registries).
- Structured audit logging via `tflog` `audit` subsystem (FR-020).
- CI tools `tools/coverage-check` (FR-021 OpenAPI coverage gate)
  and `tools/sensitive-lint` (FR-002a drift detection).
- Integration tests covering the no-`InsecureSkipVerify` invariant
  (FR-003c), the no-external-secret-store invariant (FR-001b),
  documentation coverage (FR-029), and exported-symbol doc-comment
  coverage (SC-007).

### CI / Release pipeline

- `.goreleaser.yml`: build matrix narrowed to an explicit
  `targets:` list (`linux_amd64`, `linux_arm64`, `darwin_amd64`,
  `darwin_arm64`, `windows_amd64`) — replaces the earlier
  `goos × goarch + ignore` Cartesian product.
- `.github/workflows/release.yml`:
  - Added `workflow_dispatch` trigger so a failed release can be
    rerun without re-tagging.
  - Header comment refreshed to document the 5-target matrix and
    the gating model (govulncheck strict, osv-scanner
    HIGH/CRITICAL only).
  - Bumped `GO_VERSION` to `1.26.x` (was `1.24`) to satisfy
    `go.mod`'s `toolchain` directive and remove the
    `file requires newer Go version` failure surfaced by
    govulncheck on the older runtime.
  - **Cosign step**: consolidated two duplicate `env:` blocks on
    the `Sign release archives` step into a single block; the
    duplicate keys had been rejected by GitHub's workflow parser
    with `'env' is already defined`. Hardened
    `gh release download` with `mkdir -p dist` and
    `--pattern '*.zip'`.
  - **FR-032a vulnerability gate** rewritten:
    - `govulncheck ./...` is the strict reachability gate (any
      reachable Go vulnerability fails the run).
    - `osv-scanner scan source` now runs with `--call-analysis=go`
      and writes JSON; a small Python post-processor then HARD-FAILS
      only on `database_specific.severity ∈ {HIGH, CRITICAL}` and
      drops entries marked as not-called by call-graph analysis.
      Everything else is logged and uploaded as the
      `osv-scanner-report` artifact (30-day retention).
    - New `osv-scanner.toml` at the repo root carries
      per-vulnerability suppressions (id + reason + `ignoreUntil`),
      letting accepted-risk findings be silenced without editing
      the workflow.
- `.github/workflows/release-snapshot.yml` (new): runs
  `goreleaser release --snapshot --clean --skip=publish,sign` on
  every PR/`main` push, asserts each of the five expected zips is
  produced, and uploads the snapshot as a 7-day artifact for
  debugging. GPG / cosign / SLSA paths are deliberately not
  exercised here.

[Unreleased]: https://github.com/nutanix/terraform-provider-nc2/compare/HEAD...HEAD
