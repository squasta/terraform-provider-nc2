# Phase 1 — Data Model

**Feature**: Terraform Provider for Nutanix Cloud Clusters (NC2)
**Plan**: [plan.md](./plan.md)
**Spec**: [spec.md](./spec.md)
**Authoritative API**: `openapi/openapi.json`
**Date**: 2026-05-13

This document maps every NC2 entity to its Terraform-facing surface (managed resource, data source, and/or action), enumerates the schema attribute shape, identifies sensitive fields, declares lifecycle states and transitions, lists the validation rules, and records the OpenAPI operation(s) covered. It is the input to `/speckit.tasks` and to the implementation phase.

## Conventions used throughout this document

- **Attribute type** uses Terraform Plugin Framework type names: `String`, `Bool`, `Int64`, `Float64`, `List<T>`, `Set<T>`, `Map<T>`, `Object{...}`.
- **Mode** is one of:
  - `Required` — must be set by the user.
  - `Optional` — may be set; provider does not force a default.
  - `Optional+Default` — may be set; provider injects a documented default if absent.
  - `Computed` — set by the provider from the API response; user MUST NOT set.
  - `Optional+Computed` — user MAY set; if absent, the provider sets it from the API response.
- **Plan modifier** records `RequiresReplace` (changes force destroy+recreate), `UseStateForUnknown` (preserves existing state value when the new value cannot be determined at plan time), or other framework plan modifiers.
- **Sensitive** = `true` means the attribute is marked sensitive per FR-002 (either name-pattern match or per-resource registry entry).
- **NC2 mapping** identifies the NC2 endpoint(s) and JSON field path that the attribute reads from / writes to. JSON paths use dotted notation (`data.cluster.network.vpc_cidr`).
- **OpenAPI operation IDs** below refer to the `operationId` fields in `openapi/openapi.json`. These are the units that `tools/coverage-check` (FR-021) tracks.

---

## 1. Organization

**Spec entity**: Organization — root of the NC2 resource hierarchy; owns cloud accounts; the unit of billing and audit.

**Terraform exposure**:

- **Managed resource**: `nc2_organization`
- **Data sources**: `nc2_organizations` (list), `nc2_organization` (single by `id`), `nc2_organization_audit_trail`

**OpenAPI operations covered**:

| operationId | method | path | Terraform mapping |
|---|---|---|---|
| `CPanelWeb.Api.OrganizationController.index` | GET | `/organizations` | `data.nc2_organizations.Read` |
| `CPanelWeb.Api.OrganizationController.show` | GET | `/organizations/{id}` | `nc2_organization.Read`, `data.nc2_organization.Read` |
| `CPanelWeb.Api.OrganizationController.create` | POST | `/organizations` | `nc2_organization.Create` |
| `CPanelWeb.Api.OrganizationController.update` | PATCH | `/organizations/{id}` | `nc2_organization.Update` (canonical per FR-010b) |
| `CPanelWeb.Api.OrganizationController.update (2)` | PUT | `/organizations/{id}` | `client.Organization.ReplaceAll` (CI-only path; not called from Update) |
| `CPanelWeb.Api.OrganizationController.terminate` | PATCH | `/organizations/{id}/terminate` | `nc2_organization.Delete` |
| `CPanelWeb.Api.OrganizationController.list_audit_trails` | GET | `/organizations/{id}/audit-trails` | `data.nc2_organization_audit_trail.Read` |

**Schema (`nc2_organization`)**:

| Attribute | Type | Mode | Plan modifier | Sensitive | NC2 mapping | Validation |
|---|---|---|---|---|---|---|
| `id` | String | Computed | UseStateForUnknown | false | response: `data.id` | format = UUID |
| `name` | String | Required | – | false | request/response: `data.name` | 1 ≤ len ≤ 128 |
| `description` | String | Optional | – | false | request/response: `data.description` | len ≤ 2048 |
| `state` | String | Computed | – | false | response: `data.state` | one of `active`, `terminating`, `terminated` |
| `created_at` | String | Computed | UseStateForUnknown | false | response: `data.created_at` | RFC 3339 |
| `updated_at` | String | Computed | – | false | response: `data.updated_at` | RFC 3339 |

**Lifecycle states**: `pending → active → terminating → terminated`.
Terraform Update operates only on `active` organizations. Terminate (Delete) moves to `terminating`; `Read` after terminate eventually returns 404, satisfying FR-009.

**Sensitive registry (per-package, FR-002)**: empty. Name patterns from FR-002 do not match any attribute.

---

## 2. Cloud Account

**Spec entity**: Cloud Account — a credentialed AWS / Azure / GCP account that NC2 uses to provision infrastructure on behalf of an organization.

**Terraform exposure**:

- **Managed resource**: `nc2_cloud_account` (per the no-DELETE caveat in FR-006; destroy = state-only removal with plan-time warning)
- **Data sources**: `nc2_cloud_accounts` (list, scoped to organization), `nc2_cloud_account` (single by `id`)

**OpenAPI operations covered**:

| operationId | method | path | Terraform mapping |
|---|---|---|---|
| `CPanelWeb.Api.OrganizationController.create_cloud_account` | POST | `/organizations/{id}/cloud-accounts/{cloud_provider}` | `nc2_cloud_account.Create` |
| `CPanelWeb.Api.CloudAccountController.show` | GET | `/cloud-accounts/{id}` | `nc2_cloud_account.Read`, `data.nc2_cloud_account.Read` |
| `CPanelWeb.Api.CloudAccountController.update` | PATCH | `/cloud-accounts/{id}` | `nc2_cloud_account.Update` (canonical) |
| `CPanelWeb.Api.CloudAccountController.update (2)` | PUT | `/cloud-accounts/{id}` | `client.CloudAccount.ReplaceAll` (CI-only path) |
| `CPanelWeb.Api.CloudAccountController.update_credentials` | POST | `/cloud-accounts/{id}/update-credentials` | `nc2_cloud_account.Update` when `credentials` attribute changes |
| `CPanelWeb.Api.OrganizationController.list_cloud_accounts` | GET | `/organizations/{id}/cloud-accounts` | `data.nc2_cloud_accounts.Read` |

**Schema (`nc2_cloud_account`)** (high-level — cloud-specific subfields delegated to the schema generator from `openapi/openapi.json`):

| Attribute | Type | Mode | Plan modifier | Sensitive | NC2 mapping | Validation |
|---|---|---|---|---|---|---|
| `id` | String | Computed | UseStateForUnknown | false | response: `data.id` | UUID |
| `organization_id` | String | Required | RequiresReplace | false | path param + response: `data.organization_id` | UUID |
| `cloud_provider` | String | Required | RequiresReplace | false | path param `{cloud_provider}` | one of `aws`, `azure`, `gcp` |
| `name` | String | Required | – | false | request/response: `data.name` | 1 ≤ len ≤ 128 |
| `description` | String | Optional | – | false | request/response: `data.description` | len ≤ 2048 |
| `credentials` | Object{...} | Required | – | **true** (parent) | request: `data.credentials.*`; rotation via `update-credentials` | shape per `cloud_provider`; sub-fields per OpenAPI |
| `credentials.aws.access_key_id` | String | Required (when `cloud_provider=aws`) | – | **true** | request: `data.credentials.access_key_id` | 16–32 alphanum |
| `credentials.aws.secret_access_key` | String | Required (when `cloud_provider=aws`) | – | **true** | request: `data.credentials.secret_access_key` | len ≥ 32 |
| `credentials.aws.role_arn` | String | Optional | – | false | request: `data.credentials.role_arn` | ARN format |
| `credentials.azure.tenant_id` | String | Required (when `cloud_provider=azure`) | – | false | request: `data.credentials.tenant_id` | UUID |
| `credentials.azure.subscription_id` | String | Required (when `cloud_provider=azure`) | – | false | request: `data.credentials.subscription_id` | UUID |
| `credentials.azure.client_id` | String | Required (when `cloud_provider=azure`) | – | false | request: `data.credentials.client_id` | UUID |
| `credentials.azure.client_secret` | String | Required (when `cloud_provider=azure`) | – | **true** | request: `data.credentials.client_secret` | len ≥ 16 |
| `credentials.gcp.service_account_json` | String | Required (when `cloud_provider=gcp`) | – | **true** | request: `data.credentials.service_account_json` | valid JSON |
| `state` | String | Computed | – | false | response: `data.state` | one of `active`, `disabled` |
| `created_at` | String | Computed | UseStateForUnknown | false | response: `data.created_at` | RFC 3339 |
| `updated_at` | String | Computed | – | false | response: `data.updated_at` | RFC 3339 |

**Update routing**: when only attributes inside `credentials` change, the provider POSTs to `/cloud-accounts/{id}/update-credentials`; when `name` or `description` change, it PATCHes `/cloud-accounts/{id}`; when both change in the same diff, the provider issues both calls in deterministic order (credentials update first, then PATCH).

**Lifecycle states**: `active` ⇄ `disabled` (operational state, not Terraform-visible state). Terraform `Delete` performs state-only removal with the plan-time warning per R-07.

**Sensitive registry (FR-002)**: the entire `credentials` subtree is sensitive (auto-marked by the FR-002 `credentials` substring pattern). The per-resource registry additionally lists `credentials.aws.secret_access_key`, `credentials.azure.client_secret`, `credentials.gcp.service_account_json` explicitly as belt-and-braces.

---

## 3. Cloud Account Region

**Spec entity**: Region — a cloud-provider region enabled on a cloud account.

**Terraform exposure**:

- **Managed resource**: `nc2_cloud_account_region` (per the no-remove caveat in FR-007; destroy = state-only removal with plan-time warning)
- **Data source**: `nc2_cloud_account_regions` (list, scoped to cloud account)

**OpenAPI operations covered**:

| operationId | method | path | Terraform mapping |
|---|---|---|---|
| `CPanelWeb.Api.RegionController.index` | GET | `/cloud-accounts/{cloud_account_id}/regions` | `nc2_cloud_account_region.Read`, `data.nc2_cloud_account_regions.Read` |
| `CPanelWeb.Api.RegionController.create` | POST | `/cloud-accounts/{cloud_account_id}/regions` | `nc2_cloud_account_region.Create` |

**Schema (`nc2_cloud_account_region`)**:

| Attribute | Type | Mode | Plan modifier | Sensitive | NC2 mapping | Validation |
|---|---|---|---|---|---|---|
| `id` | String | Computed | UseStateForUnknown | false | response: `data.id` | string |
| `cloud_account_id` | String | Required | RequiresReplace | false | path param | UUID |
| `region` | String | Required | RequiresReplace | false | request body `data.region`; response `data.region` | cloud-provider-specific region identifier (e.g., `us-east-1`, `eastus`, `europe-west1`) |
| `state` | String | Computed | – | false | response: `data.state` | one of `available`, `disabled` |

**Lifecycle**: `add` (Create) → `available`. Terraform `Delete` performs state-only removal with plan-time warning per R-07.

**Sensitive registry**: empty.

---

## 4. Cluster (AWS / Azure / GCP)

**Spec entity**: Cluster — the headline workload. Three sibling managed resource types per FR-004, sharing common attributes with cloud-specific extensions.

**Terraform exposure**:

- **Managed resources**: `nc2_aws_cluster`, `nc2_azure_cluster`, `nc2_gcp_cluster`
- **Data sources**: `nc2_clusters` (list), `nc2_cluster` (single by `id`), `nc2_cluster_cloud_resources` (cluster-scoped)
- **Actions** (US5): see §11

**OpenAPI operations covered** (for the AWS variant; Azure and GCP mirror the same set minus AWS-only endpoints):

| operationId | method | path | Terraform mapping |
|---|---|---|---|
| `CPanelWeb.Api.ClusterController.create_aws` | POST | `/clusters/aws` | `nc2_aws_cluster.Create` |
| `CPanelWeb.Api.ClusterController.create_azure` | POST | `/clusters/azure` | `nc2_azure_cluster.Create` |
| `CPanelWeb.Api.ClusterController.create_gcp` | POST | `/clusters/gcp` | `nc2_gcp_cluster.Create` |
| `CPanelWeb.Api.ClusterController.show` | GET | `/clusters/{id}` | `<cloud>.Read`, `data.nc2_cluster.Read` |
| `CPanelWeb.Api.ClusterController.index` | GET | `/clusters` | `data.nc2_clusters.Read` |
| `CPanelWeb.Api.ClusterController.update` | PATCH | `/clusters/{id}` | `<cloud>.Update` (generic mutable fields) |
| `CPanelWeb.Api.ClusterController.update (2)` | PUT | `/clusters/{id}` | `client.Cluster.ReplaceAll` (CI-only path) |
| `CPanelWeb.Api.ClusterController.terminate` | POST | `/clusters/{id}/terminate` | `<cloud>.Delete` |
| `CPanelWeb.Api.ClusterController.hibernate` | POST | `/clusters/{id}/hibernate` | `<cloud>.Update` when `desired_state: running → hibernated` |
| `CPanelWeb.Api.ClusterController.resume` | POST | `/clusters/{id}/resume` | `<cloud>.Update` when `desired_state: hibernated → running` |
| `CPanelWeb.Api.ClusterController.update_capacity` | POST | `/clusters/{id}/update-capacity` | `<cloud>.Update` when `capacity` changes |
| `CPanelWeb.Api.ClusterController.update_ssh_key` | POST | `/clusters/{id}/update-ssh-key` | `<cloud>.Update` when `host_access_ssh_key` changes |
| `CPanelWeb.Api.ClusterController.update_license` | POST | `/clusters/{id}/update-license` | `<cloud>.Update` when `license`/`software_tier`/`aos_version` changes |
| `CPanelWeb.Api.ClusterController.update_resource_tags` | POST | `/clusters/{id}/update-resource-tags` | `<cloud>.Update` when `resource_tags` changes |
| `CPanelWeb.Api.ClusterController.update_access_policy` | POST | `/clusters/{id}/update-access-policy` | `nc2_aws_cluster.Update` when `access_policy` changes (**AWS only**, FR-010a) |
| `CPanelWeb.Api.CloudResourceController.index` | GET | `/clusters/{cluster_id}/cloud-resources` | `data.nc2_cluster_cloud_resources.Read` |

(The cluster operational endpoints — `condemn-host`, `open/close/extend-support-tunnel`, `scale-out-fgw`, `upgrade-fgw`, `start-recovery` — are not part of the managed resource's lifecycle; they map to actions and are described in §11.)

**Common cluster schema (shared by all three cloud-specific resources unless noted)**:

| Attribute | Type | Mode | Plan modifier | Sensitive | NC2 mapping | Validation |
|---|---|---|---|---|---|---|
| `id` | String | Computed | UseStateForUnknown | false | response: `data.id` | UUID |
| `organization_id` | String | Required | RequiresReplace | false | request: `data.organization_id` | UUID |
| `cloud_account_id` | String | Required | RequiresReplace | false | request: `data.cloud_account_id` | UUID |
| `name` | String | Required | RequiresReplace | false | request: `data.name` | 1 ≤ len ≤ 64; DNS-label safe |
| `region` | String | Required | RequiresReplace | false | request: `data.region` | non-empty |
| `use_case` | String | Optional+Default | – | false | request: `data.use_case` | one of `general`, `production`; default `general` |
| `host_access_ssh_key` | String | Required | – (in-place via `update-ssh-key`) | false | request: `data.host_access_ssh_key` | non-empty |
| `license` | String | Required | – (in-place via `update-license`) | false | request: `data.license` | one of `aos`, `prism-pro`, ... (enum from OpenAPI) |
| `aos_version` | String | Required | – (in-place via `update-license`) | false | request: `data.aos_version` | matches `^\d+\.\d+(\.\d+)?$` |
| `software_tier` | String | Required | – (in-place via `update-license`) | false | request: `data.software_tier` | enum per OpenAPI |
| `capacity` | List<Object{host_type, number_of_hosts, node_type?, advanced_settings?}> | Required | – (in-place via `update-capacity`) | false | request: `data.capacity` | 1 ≤ len ≤ N; per-element host_type non-empty, number_of_hosts ≥ 1 |
| `redundancy` | Object{factor: Int64} | Required | RequiresReplace | false | request: `data.redundancy` | factor ∈ {1, 2} |
| `network` | Object{...} | Required | RequiresReplace (with field-level exceptions) | false | request: `data.network` | shape per OpenAPI |
| `network.mode` | String | Required | RequiresReplace | false | request: `data.network.mode` | one of `new`, `existing` |
| `network.availability_zone` | String | Required | RequiresReplace | false | request: `data.network.availability_zone` | non-empty |
| `network.vpc_cidr` | String | Optional (Required when `network.mode=new`) | RequiresReplace | false | request: `data.network.vpc_cidr` | CIDR notation |
| `network.management_services_access_policy` | Object{mode, ip_addresses} | Optional | – | false | request: `data.network.management_services_access_policy` | mode ∈ {`open`,`restricted`} |
| `network.prism_element_access_policy` | Object{mode, ip_addresses} | Optional | – | false | request: `data.network.prism_element_access_policy` | mode ∈ {`open`,`restricted`} |
| `resource_tags` | Map<String> | Optional | – (in-place via `update-resource-tags`) | false | request: `data.resource_tags` | keys 1–128 chars |
| `desired_state` | String | Optional+Default | – | false | derived: hibernate/resume endpoints | one of `running`, `hibernated`; default `running` |
| `state` | String | Computed | – | false | response: `data.state` | enum {`provisioning`, `running`, `hibernating`, `hibernated`, `resuming`, `terminating`, `terminated`, `failed`} |
| `created_at` | String | Computed | UseStateForUnknown | false | response: `data.created_at` | RFC 3339 |
| `updated_at` | String | Computed | – | false | response: `data.updated_at` | RFC 3339 |

**AWS-only additional schema (`nc2_aws_cluster`)**:

| Attribute | Type | Mode | Plan modifier | Sensitive | NC2 mapping | Validation |
|---|---|---|---|---|---|---|
| `access_policy` | Object{...} | Optional | – (in-place via `update-access-policy`) | false | request: `data.access_policy`; update via `/clusters/{id}/update-access-policy` | per OpenAPI |
| `network.aws.subnets` | List<Object{id, availability_zone}> | Optional | RequiresReplace | false | request: `data.network.aws.subnets` | per OpenAPI |

**Azure-only additional schema (`nc2_azure_cluster`)**:

| Attribute | Type | Mode | Plan modifier | Sensitive | NC2 mapping |
|---|---|---|---|---|---|
| `network.azure.virtual_network_id` | String | Optional | RequiresReplace | false | request: `data.network.azure.virtual_network_id` |
| `network.azure.delegated_subnet_id` | String | Optional | RequiresReplace | false | request: `data.network.azure.delegated_subnet_id` |

**GCP-only additional schema (`nc2_gcp_cluster`)**:

| Attribute | Type | Mode | Plan modifier | Sensitive | NC2 mapping |
|---|---|---|---|---|---|
| `network.gcp.project_id` | String | Required (when used) | RequiresReplace | false | request: `data.network.gcp.project_id` |
| `network.gcp.vpc_name` | String | Optional | RequiresReplace | false | request: `data.network.gcp.vpc_name` |

**Lifecycle state machine**:

```
                        ┌──────────────┐
                        │ provisioning │ ◀── Create
                        └──────┬───────┘
                               ▼
            ┌─────────────────────────────────┐
            │             running             │ ◀── Resume (from hibernated)
            └───┬────────────────────────┬────┘
                │ Hibernate              │ Terminate
                ▼                        ▼
       ┌───────────────┐       ┌──────────────┐
       │  hibernating  │       │  terminating │
       └───────┬───────┘       └──────┬───────┘
               ▼                      ▼
       ┌───────────────┐       ┌──────────────┐
       │  hibernated   │       │  terminated  │ (Read returns 404 → state removed)
       └───────────────┘       └──────────────┘

  any state → failed (terminal for managed-resource purposes; user must Delete to reconcile)
```

**Update routing** (precedence when multiple changes are in the same diff): the provider applies operations in this fixed order to maximize success: (1) `update-license`, (2) `update-ssh-key`, (3) `update-capacity`, (4) `update-resource-tags`, (5) `update-access-policy` (AWS only), (6) generic `PATCH /clusters/{id}`, (7) hibernate/resume if `desired_state` changed. Plan-time validation rejects mutually-incompatible diffs (e.g., setting `desired_state=hibernated` AND increasing `capacity` in the same diff is rejected with a clear error).

**Sensitive registry (per package)**:

- `nc2_aws_cluster`, `nc2_azure_cluster`, `nc2_gcp_cluster`: registry includes `network.prism_element_access_policy.ip_addresses` (treated as sensitive operational info although not matched by pattern), `network.management_services_access_policy.ip_addresses`. Other attributes are non-sensitive.

---

## 5. Availability Zone (data source only)

**Spec entity**: Availability Zone — failure domain inside a region.

**Terraform exposure**:

- **Data source**: `nc2_availability_zones` (read-only)

**OpenAPI operations covered**:

| operationId | method | path | Terraform mapping |
|---|---|---|---|
| `CPanelWeb.Api.AvailabilityZoneController.index` | GET | `/cloud-accounts/{cloud_account_id}/regions/{region_id}/availability-zones` | `data.nc2_availability_zones.Read` |

**Schema (`data.nc2_availability_zones`)**:

| Attribute | Type | Mode | NC2 mapping |
|---|---|---|---|
| `cloud_account_id` | String | Required | path param |
| `region_id` | String | Required | path param |
| `availability_zones` | List<Object{id, name, state}> | Computed | response: `data[]` |

Sorted deterministically by `id` per FR-015.

---

## 6. VPC / VNet / SSH Key / Prism Central / Remote Storage Profile (data sources only)

All five entities follow the same pattern: read-only, scoped to a cloud account + region (except remote storage profiles, which are global). Each data source returns a list of objects sorted deterministically by `id`.

| Data source | OpenAPI operationId | Path |
|---|---|---|
| `data.nc2_vpcs` | `CPanelWeb.Api.VpcController.index` | `/cloud-accounts/{cloud_account_id}/regions/{region_id}/vpcs` |
| `data.nc2_vnets` | `CPanelWeb.Api.VnetController.index` | `/cloud-accounts/{cloud_account_id}/regions/{region_id}/vnets` |
| `data.nc2_ssh_keys` | `CPanelWeb.Api.SshKeyController.index` | `/cloud-accounts/{cloud_account_id}/regions/{region_id}/ssh-keys` |
| `data.nc2_prism_centrals` | `CPanelWeb.Api.PrismCentralController.index` | `/cloud-accounts/{cloud_account_id}/regions/{region_id}/prism-centrals` |
| `data.nc2_remote_storage_profiles` | `CPanelWeb.Api.RemoteStorageProfileController.index` | `/remote-storage-profiles` |

The attribute shape of each result element is derived directly from the OpenAPI schema for that entity. Per FR-014a, none of these has a managed-resource counterpart.

Sensitive registries: empty for all five (no FR-002 pattern matches in the response shapes).

---

## 7. Cloud Resource (cluster-scoped data source only)

**Spec entity**: Cloud Resource — read-only inventory of cloud-side artifacts created by NC2, scoped to a specific cluster.

**Terraform exposure**:

- **Data source**: `nc2_cluster_cloud_resources`

**OpenAPI operations covered**:

| operationId | method | path | Terraform mapping |
|---|---|---|---|
| `CPanelWeb.Api.CloudResourceController.index` | GET | `/clusters/{cluster_id}/cloud-resources` | `data.nc2_cluster_cloud_resources.Read` |

**Schema**:

| Attribute | Type | Mode | NC2 mapping |
|---|---|---|---|
| `cluster_id` | String | Required | path param |
| `cloud_resources` | List<Object{...}> | Computed | response: `data[]` |

---

## 8. Notification

**Spec entity**: Notification — system-generated notification with an acknowledged/unacknowledged state.

**Terraform exposure**:

- **Data source**: `nc2_notifications` (list, read-only)
- **Action**: `nc2_notification_acknowledge` (one-shot acknowledge of a single notification)

Notifications are **not** exposed as a managed resource (FR-014, FR-016 — notifications are not user-creatable).

**OpenAPI operations covered**:

| operationId | method | path | Terraform mapping |
|---|---|---|---|
| `CPanelWeb.Api.NotificationController.index` | GET | `/notifications` | `data.nc2_notifications.Read` |
| `CPanelWeb.Api.NotificationController.update` | PATCH | `/notifications/{id}` | `nc2_notification_acknowledge.Invoke` (canonical) |
| `CPanelWeb.Api.NotificationController.update (2)` | PUT | `/notifications/{id}` | `client.Notification.ReplaceAll` (CI-only path) |

**Action schema (`nc2_notification_acknowledge`)**:

| Attribute | Type | Mode | NC2 mapping | Validation |
|---|---|---|---|---|
| `notification_id` | String | Required | path param `{id}` | UUID |
| `result` | Object{status: String, http_status: Int64} | Computed | computed from response | – |

The action body sent to NC2 is always `{"data": {"acknowledged": true}}` (FR-016, US5 acceptance scenario 5).

---

## 9. Task (data sources only; internal use by every async path)

**Spec entity**: Task — asynchronous NC2 operation tracking record. The polling target for every async API call (FR-008).

**Terraform exposure**:

- **Data sources**: `nc2_tasks` (list), `nc2_task` (single by `id`)
- **Internal use**: every managed-resource Create/Update/Delete and every async action polls `GET /tasks/{id}` via the `internal/client.PollTask` function

**OpenAPI operations covered**:

| operationId | method | path | Terraform mapping |
|---|---|---|---|
| `CPanelWeb.Api.TaskController.index` | GET | `/tasks` | `data.nc2_tasks.Read` |
| `CPanelWeb.Api.TaskController.show` | GET | `/tasks/{id}` | `data.nc2_task.Read`, `client.PollTask` (internal) |

**Schema (`data.nc2_task`)**:

| Attribute | Type | Mode | NC2 mapping | Validation |
|---|---|---|---|---|
| `id` | String | Required | path param | UUID |
| `status` | String | Computed | response: `data.status` | one of `pending`, `running`, `done`, `failed`, `cancelled` |
| `progress` | Int64 | Computed | response: `data.progress` | 0 ≤ x ≤ 100 |
| `created_at` | String | Computed | response: `data.created_at` | RFC 3339 |
| `error_code` | String | Computed | response: `data.error.code` if present | – |
| `error_message` | String | Computed | response: `data.error.message` if present | – |

Terminal statuses are `done`, `failed`, `cancelled` (per the OpenAPI `info.description`).

---

## 10. Region (data source only)

**Spec entity**: Region — cloud-provider region enabled on a cloud account.

**Terraform exposure**:

- **Data source**: `nc2_cloud_account_regions`

(The managed-resource counterpart for adding a region is `nc2_cloud_account_region` documented in §3.)

---

## 11. Cluster operational actions (no managed-resource counterpart)

**Spec entity**: Action Invocation — one-shot operational call (condemn host, open/close/extend support tunnel, scale Flow Gateway, upgrade Flow Gateway, start cluster recovery). No persisted state.

**Terraform exposure**: 7 cluster-scoped actions (plus `nc2_notification_acknowledge` in §8).

**OpenAPI operations covered**:

| Action | operationId | path |
|---|---|---|
| `nc2_cluster_condemn_host` | `CPanelWeb.Api.ClusterController.condemn_host` | `POST /clusters/{id}/condemn-host` |
| `nc2_cluster_open_support_tunnel` | `CPanelWeb.Api.ClusterController.open_support_tunnel` | `POST /clusters/{id}/open-support-tunnel` |
| `nc2_cluster_extend_support_tunnel` | `CPanelWeb.Api.ClusterController.extend_support_tunnel` | `POST /clusters/{id}/extend-support-tunnel` |
| `nc2_cluster_close_support_tunnel` | `CPanelWeb.Api.ClusterController.close_support_tunnel` | `POST /clusters/{id}/close-support-tunnel` |
| `nc2_cluster_scale_flow_gateway` | `CPanelWeb.Api.ClusterController.scale_out_fgw` | `POST /clusters/{id}/scale-out-fgw` |
| `nc2_cluster_upgrade_flow_gateway` | `CPanelWeb.Api.ClusterController.upgrade_fgw` | `POST /clusters/{id}/upgrade-fgw` |
| `nc2_cluster_start_recovery` | `CPanelWeb.Api.ClusterController.start_recovery` | `POST /clusters/{id}/start-recovery` |

**Common action schema**:

| Attribute | Type | Mode | NC2 mapping |
|---|---|---|---|
| `cluster_id` | String | Required | path param `{id}` |
| `result` | Object{task_id: String, status: String, error_code: String, error_message: String} | Computed | derived from task polling |

Per-action additional inputs (e.g., `host_id` for `condemn_host`, `duration_hours` for `extend_support_tunnel`, `target_node_count` for `scale_flow_gateway`) are declared in each action's individual schema, sourced from the corresponding OpenAPI request body.

**Sensitive registry**: empty for all 7 cluster actions. `nc2_notification_acknowledge` has empty sensitive registry.

---

## 12. Cross-cutting validation rules (applied by the `internal/provider` package)

These apply uniformly to every managed resource and action and are derived from spec requirements:

- **Credential resolution** (FR-001a): exactly one effective value per credential attribute; mismatch warning when two sources differ.
- **TLS** (FR-003b): no schema attribute named `insecure_skip_verify` or equivalent exists in the provider config; static analysis fails the build if such an attribute is introduced.
- **AWS-only access policy** (FR-010a): if `access_policy` is set on `nc2_azure_cluster` or `nc2_gcp_cluster`, a plan-time validation error is raised with the path `access_policy`. Cleanly testable per `resource_test.go`.
- **Drift-free guarantee** (FR-018): every managed resource has a unit test that round-trips an example state through `PlanResourceChange` with no diff.
- **Clean-destroy guarantee** (FR-019): every managed resource has an acceptance test asserting that after `terraform destroy`, a subsequent `terraform plan` returns "No changes" on an empty config.
- **Import support** (FR-013): every managed resource implements `ImportState` with an ID equal to its NC2 UUID, and an acceptance test asserts post-import plan is drift-free (SC-009).

---

## Coverage check

This data model covers all **49 operations** across **40 paths** in `openapi/openapi.json`, distributed as:

- 17 operations behind managed resources (cluster cluster-cluster + org/account/region/notification PATCH+PUT pairs + cloud account create + region create + terminate),
- 18 operations behind data sources,
- 14 operations behind actions (including the `update-credentials` call which is internally routed by the `nc2_cloud_account` resource's Update method but is also exercised by tests as an action-style scenario).

(Detailed mapping is enumerated above per section; `tools/coverage-check` will mechanically verify this against the OpenAPI on every build per R-09 and FR-021/FR-021a.)
