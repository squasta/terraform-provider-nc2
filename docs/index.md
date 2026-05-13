# Nutanix Cloud Clusters (NC2) Provider

The `nc2` provider lets you manage Nutanix Cloud Clusters (NC2)
organizations, cloud accounts, regions, clusters, and operational
actions through Terraform.

## Quick start

```terraform
terraform {
  required_providers {
    nc2 = {
      source  = "nutanix/nc2"
      version = "~> 0.1"
    }
  }
}

provider "nc2" {
  # Credentials may be supplied via the provider block, environment
  # variables (NC2_API_KEY, NC2_KEY_ID, NC2_ISSUER), or a credentials
  # file (default ~/.nc2/credentials). See the authentication guide.
}
```

```terraform
resource "nc2_organization" "demo" {
  name        = "demo-org"
  description = "Demo organization."
}

resource "nc2_cloud_account" "aws" {
  organization_id = nc2_organization.demo.id
  cloud_provider  = "aws"
  name            = "demo-aws"
  credentials = {
    access_key_id     = var.aws_access_key_id
    secret_access_key = var.aws_secret_access_key
  }
}
```

## Guides

- [Getting started](./guides/getting-started.md) — installation,
  first apply.
- [Authentication](./guides/authentication.md) — credential
  resolution rules (FR-001a) and the deliberate absence of native
  external secret-store imports (FR-001b).
- [Security hardening](./guides/security-hardening.md) — TLS
  posture (FR-003b/c), sensitive-field redaction (FR-002), audit
  log routing (FR-020c), release signing (FR-032).
- [Async operations & tasks](./guides/async-and-tasks.md) — task
  polling, timeouts, troubleshooting via task IDs and audit records.

## Resources

- [`nc2_organization`](./resources/nc2_organization.md)
- [`nc2_cloud_account`](./resources/nc2_cloud_account.md)
- [`nc2_cloud_account_region`](./resources/nc2_cloud_account_region.md)
- [`nc2_aws_cluster`](./resources/nc2_aws_cluster.md)
- [`nc2_azure_cluster`](./resources/nc2_azure_cluster.md)
- [`nc2_gcp_cluster`](./resources/nc2_gcp_cluster.md)

## Data Sources

- [`nc2_organizations`](./data-sources/nc2_organizations.md)
- [`nc2_organization`](./data-sources/nc2_organization.md)
- [`nc2_organization_audit_trail`](./data-sources/nc2_organization_audit_trail.md)
- [`nc2_cloud_accounts`](./data-sources/nc2_cloud_accounts.md)
- [`nc2_cloud_account`](./data-sources/nc2_cloud_account.md)
- [`nc2_cloud_account_regions`](./data-sources/nc2_cloud_account_regions.md)
- [`nc2_clusters`](./data-sources/nc2_clusters.md)
- [`nc2_cluster`](./data-sources/nc2_cluster.md)
- [`nc2_cluster_cloud_resources`](./data-sources/nc2_cluster_cloud_resources.md)
- [`nc2_availability_zones`](./data-sources/nc2_availability_zones.md)
- [`nc2_vpcs`](./data-sources/nc2_vpcs.md)
- [`nc2_vnets`](./data-sources/nc2_vnets.md)
- [`nc2_ssh_keys`](./data-sources/nc2_ssh_keys.md)
- [`nc2_prism_centrals`](./data-sources/nc2_prism_centrals.md)
- [`nc2_remote_storage_profiles`](./data-sources/nc2_remote_storage_profiles.md)
- [`nc2_notifications`](./data-sources/nc2_notifications.md)
- [`nc2_tasks`](./data-sources/nc2_tasks.md)
- [`nc2_task`](./data-sources/nc2_task.md)

## Actions

- [`nc2_cluster_condemn_host`](./actions/nc2_cluster_condemn_host.md)
- [`nc2_cluster_open_support_tunnel`](./actions/nc2_cluster_open_support_tunnel.md)
- [`nc2_cluster_extend_support_tunnel`](./actions/nc2_cluster_extend_support_tunnel.md)
- [`nc2_cluster_close_support_tunnel`](./actions/nc2_cluster_close_support_tunnel.md)
- [`nc2_cluster_scale_flow_gateway`](./actions/nc2_cluster_scale_flow_gateway.md)
- [`nc2_cluster_upgrade_flow_gateway`](./actions/nc2_cluster_upgrade_flow_gateway.md)
- [`nc2_cluster_start_recovery`](./actions/nc2_cluster_start_recovery.md)
- [`nc2_notification_acknowledge`](./actions/nc2_notification_acknowledge.md)

## Security posture (summary)

- **No external secret-store dependencies** (FR-001b). Credentials
  must arrive from the provider block, environment variables, or a
  credentials file — see [Authentication](./guides/authentication.md).
- **Strict TLS verification** (FR-003c). The provider intentionally
  does not expose `insecure_skip_verify` and the codebase contains
  static checks ensuring `&tls.Config{InsecureSkipVerify: true}` is
  never constructed.
- **Sensitive-field redaction** (FR-002). Pattern-classified
  attributes (`api_key`, `secret`, etc.) and per-resource registry
  entries are redacted from plan output, state previews, and audit
  logs.
- **Structured audit logging** (FR-020). Every NC2 API call emits
  one `tflog.Info` record under the `audit` subsystem, gated by
  `TF_LOG_PROVIDER_AUDIT=INFO`.

See the [security hardening guide](./guides/security-hardening.md)
for the full threat model and operational checklist.
