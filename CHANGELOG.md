# Changelog

All notable changes to the `terraform-provider-nc2` provider will be
documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- **BREAKING (pre-1.0)**: Hibernate / resume (`desired_state`) is
  now AWS-only on `nc2_aws_cluster` (FR-012). The attribute has
  been removed from `nc2_azure_cluster` and `nc2_gcp_cluster`;
  setting `desired_state` on those resources now produces a schema
  error at plan time. The shared `clustershared.Model` no longer
  carries the `DesiredState` field — AWS embeds the new
  `clustershared.HibernatingModel` instead, which adds the
  attribute back.

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

[Unreleased]: https://github.com/nutanix/terraform-provider-nc2/compare/HEAD...HEAD
