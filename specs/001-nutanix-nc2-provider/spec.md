# Feature Specification: Terraform Provider for Nutanix Cloud Clusters (NC2)

**Feature Branch**: `001-nutanix-nc2-provider`

**Created**: 2026-05-13

**Status**: Draft

**Input**: User description: "I want to create a terraform provider for Nutanix Cloud Clusters (NC2) using the latest terraform provider framework. This should implement all API for Nutanix NC2 as documented here: https://www.nutanix.dev/api_reference/apis/nc2.html Pay particular attention to extensively test lifecycle events, create, read, update and delete must be tested extensively including in E2E with terraform plan -> apply > plan again -> delete. All code must be documented inline and in general documentation."

**Authoritative API source**: `openapi/openapi.json` (committed in this repository) — Nutanix Cloud Clusters API Reference, OpenAPI 3.0.0, version `v2`, server `https://cloud.nutanix.com/api/v2`. This file is the source of truth for endpoint coverage and schema; the public HTML reference at `https://www.nutanix.dev/api_reference/apis/nc2.html` is its public mirror.

## Clarifications

### Session 2026-05-13

- Q: Which NC2 response fields beyond the MyNutanix authentication inputs must be treated as sensitive in Terraform state, plan output, and logs? → A: Sensitive-by-default for credential-like fields. Apply a maintained name-pattern list (`*credential*`, `*password*`, `*secret*`, `*token*`, `*private_key*`, `*api_key*`, case-insensitive substring match), plus a per-resource sensitive registry in the source. Unknown new fields default to non-sensitive but trigger a CI lint warning asking the maintainer to classify them.
- Q: What TLS verification policy must the provider apply when talking to the NC2 API endpoint? → A: Strict only — TLS certificate verification is always enforced and cannot be disabled. CA trust customization is supported through the OS trust store (default) or an explicit `ca_bundle` provider attribute pointing to a PEM file. No `insecure_skip_verify` flag is exposed in any build.
- Q: Where may credentials come from, and how is precedence resolved? → A: Three-source layered model with no native external-secret-store integration. Precedence highest → lowest: provider configuration block, then `NC2_API_KEY` / `NC2_KEY_ID` / `NC2_ISSUER` environment variables, then an optional `credentials_file` (INI / TOML, profile-aware, default path `~/.nc2/credentials`). When two sources both supply a value for the same attribute and the values differ, the provider emits a warning naming both sources and uses the higher-precedence value. External secret managers (Vault, AWS SM, etc.) are consumed by users through existing Terraform data sources feeding the provider block — no provider-internal integration.
- Q: How is the provider release signed, and what supply-chain attestation rides along? → A: GPG signature (for Terraform Registry compatibility) plus Sigstore cosign keyless signatures plus SLSA Build Level 3 provenance generated in CI (e.g., via `slsa-github-generator`), plus a dependency vulnerability gate (`govulncheck` and `osv-scanner`) that fails the release pipeline on any HIGH or CRITICAL CVE in the transitive Go dependency graph.
- Q: Does the provider emit its own audit log of NC2 API calls, and how is it routed? → A: Yes — one structured JSON audit line per outbound NC2 HTTP call via Terraform's `tflog` facility at INFO level, with fields `ts`, `correlation_id`, `terraform_op`, `http_method`, `path`, `status`, `latency_ms`, `nc2_task_id` (when present), and `nc2_error_code` (when present). Sensitive fields are redacted per FR-002. Routing and persistence are delegated to Terraform's standard `TF_LOG` / `TF_LOG_PATH` environment variables; the provider exposes no private sink, file path, or external transport.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Manage NC2 organizations and cloud accounts as Terraform resources (Priority: P1) — MVP foundation

An IaC platform owner wants to manage the NC2 tenant hierarchy — organizations and cloud accounts (with credential rotation) — as Terraform resources, so that account onboarding, credential rotation, and offboarding follow the same review-and-apply workflow as the rest of the infrastructure.

**Why this priority**: Organizations and cloud accounts are the dependency root of the entire NC2 API. Clusters cannot exist without an organization plus a cloud account; every region, SSH key, VPC, VNet, availability zone, and Prism Central lookup is scoped to a cloud account. Until this layer is manageable from Terraform, no other layer can be reliably provisioned from Terraform. This is the MVP slice and a hard prerequisite for User Story 2.

**Independent Test**: Apply an `nc2_organization` resource and an `nc2_cloud_account` resource against a sandbox NC2 tenant, verify they exist via the NC2 API, rotate the cloud account credentials via Terraform and confirm the update succeeds, then run the full lifecycle — `init` → `plan` → `apply` → `plan` (drift-free) → `destroy` → `plan` (clean state) — and verify the documented destroy semantics for each resource (full terminate for organizations; state-only removal with plan-time warning for cloud accounts per FR-006).

**Acceptance Scenarios**:

1. **Given** valid administrator credentials, **When** the user applies an `nc2_organization` block, **Then** the organization is created via `POST /organizations` and a read-back via `GET /organizations/{id}` matches the input.
2. **Given** an applied organization, **When** the user runs `terraform plan` again with no configuration changes, **Then** the plan reports "No changes" (drift-free guarantee).
3. **Given** an existing organization, **When** the user applies an `nc2_cloud_account` block tied to it, **Then** the cloud account is created via `POST /organizations/{id}/cloud-accounts/{cloud_provider}` and a read-back via `GET /cloud-accounts/{id}` matches the input.
4. **Given** an existing cloud account, **When** the user rotates credentials by changing the credential attribute, **Then** the provider calls `POST /cloud-accounts/{id}/update-credentials`, plan-after-apply is clean, and the new credentials are accepted on the next operation.
5. **Given** an applied organization, **When** the user runs `terraform destroy`, **Then** the organization is terminated via `PATCH /organizations/{id}/terminate` and removed from Terraform state, and a subsequent plan reports no managed resources.
6. **Given** an applied cloud account, **When** the user runs `terraform destroy`, **Then** the provider emits a plan-time warning (per FR-006: no DELETE endpoint exists), removes the resource from state on apply without calling any NC2 mutation endpoint, and a subsequent plan reports clean state.
7. **Given** an organization or cloud account deleted out-of-band via the NC2 console, **When** the user runs `terraform plan`, **Then** the plan reports the resource as needing re-creation (correct handling of read-time `404`).

---

### User Story 2 - Manage NC2 clusters across AWS, Azure, and GCP (Priority: P2)

A platform engineer wants to declaratively manage NC2 clusters on AWS, Azure, and Google Cloud — full create / read / update / delete lifecycle, including in-place updates for the fields NC2 supports updating in place — using a dedicated resource type per cloud, with the same drift-free / clean-destroy quality bar across all three.

**Why this priority**: This is the headline value of the provider, but it strictly depends on User Story 1 (organizations + cloud accounts) being in place — a cluster cannot exist without them, and the user explicitly requested that the tenant-hierarchy work be done first and cluster work follow. All three clouds ship together because the user requested coverage of "all API for Nutanix NC2 as documented"; once the org/account foundation exists, there is no reason to delay Azure / GCP behind AWS.

**Independent Test**: For each of AWS, Azure, and GCP, write a Terraform configuration containing one cluster resource against a sandbox NC2 cloud account (provisioned via User Story 1); run the full lifecycle `init` → `plan` → `apply` → `plan` (drift-free) → in-place update → `plan` (drift-free) → `destroy` → `plan` (clean state); verifiable independently per cloud since each is its own resource type (`nc2_aws_cluster`, `nc2_azure_cluster`, `nc2_gcp_cluster`).

**Acceptance Scenarios** — cluster create / read / destroy on every cloud:

1. **Given** a valid cloud account and an enabled region for the target cloud, **When** the user writes an `nc2_aws_cluster` / `nc2_azure_cluster` / `nc2_gcp_cluster` block with the minimum required fields and runs `terraform plan`, **Then** the plan reports exactly one resource to add with the configured attributes shown.
2. **Given** that plan, **When** the user runs `terraform apply`, **Then** the cluster is created via `POST /clusters/{aws|azure|gcp}`, the async task is polled to completion per FR-008, the cluster reaches a running state in NC2, and Terraform records its identifier in state.
3. **Given** an applied cluster with unchanged configuration on any of the three clouds, **When** the user runs `terraform plan` again, **Then** the plan reports "No changes" (drift-free guarantee across all clouds — see FR-018, SC-002).
4. **Given** an applied cluster, **When** the user runs `terraform destroy`, **Then** the cluster is terminated via `POST /clusters/{id}/terminate`, removed from Terraform state, and a subsequent plan reports no managed resources.
5. **Given** an applied cluster deleted out-of-band via the NC2 console, **When** the user runs `terraform plan`, **Then** the plan reports the resource as needing re-creation (correct handling of read-time `404`).

**Acceptance Scenarios** — cluster in-place updates:

6. **Given** an applied cluster, **When** the user changes capacity, **Then** `terraform plan` shows an in-place update and `apply` uses `POST /clusters/{id}/update-capacity` without recreating the cluster.
7. **Given** an applied cluster, **When** the user changes the SSH key, license, or resource tags, **Then** plan/apply use the corresponding `update-ssh-key` / `update-license` / `update-resource-tags` endpoint in place.
8. **Given** an applied `nc2_aws_cluster`, **When** the user changes the access policy, **Then** plan/apply use `POST /clusters/{id}/update-access-policy` (AWS-only path per FR-010a). The same change attempted on `nc2_azure_cluster` or `nc2_gcp_cluster` MUST be rejected at plan time.
9. **Given** an applied cluster, **When** the user attempts to change a field NC2 does not allow updating in place (e.g., region), **Then** `terraform plan` shows a destroy + recreate rather than an in-place update.
10. **Given** any successful in-place update, **When** the user runs `terraform plan` immediately afterward, **Then** plan reports no further changes.

**Acceptance Scenarios** — cross-cloud parity:

11. **Given** the three cloud-specific cluster resources, **When** a user inspects their attribute schemas, **Then** common attributes (name, organization, cloud account, region, capacity, tags, redundancy, network mode) share consistent naming while cloud-specific attributes are scoped to that cloud's resource only.

---

### User Story 3 - Inventory and discovery via data sources (Priority: P2)

A platform engineer wants to look up NC2 inventory — regions enabled on a cloud account, availability zones, VPCs/VNets, SSH keys, Prism Centrals, remote storage profiles, cluster-scoped cloud resources, notifications, tasks — without making manual API calls, so that cluster definitions can reference live inventory by name or filter rather than hard-coded identifiers.

**Why this priority**: Without data sources, users must hard-code IDs that drift over time. Data sources make configurations portable and readable. They are also the natural dependency input for User Story 2 cluster resources (e.g., picking a region or VPC for a cluster). They can be developed in parallel with User Story 2 because they are read-only and share no mutable state.

**Independent Test**: For each data source, write a Terraform configuration that reads it with no managed resources, run `terraform plan` / `terraform refresh`, and verify the returned values match what the NC2 API returns for the same query parameters.

**Acceptance Scenarios**:

1. **Given** a configured cloud account, **When** the user references `data "nc2_cloud_account_regions"`, **Then** the data source returns the regions enabled on that cloud account.
2. **Given** a region, **When** the user references `data "nc2_availability_zones"`, **Then** the data source returns the AZs available in that region.
3. **Given** a cloud account and region, **When** the user references data sources for VPCs, VNets, SSH keys, and Prism Centrals, **Then** each returns results consistent with the NC2 API for the same parameters.
4. **Given** any deployment, **When** the user references `data "nc2_remote_storage_profiles"`, `data "nc2_notifications"`, `data "nc2_tasks"`, or `data "nc2_cluster_cloud_resources"` (cluster-scoped), **Then** each returns results consistent with the NC2 API.
5. **Given** any list-style data source, **When** queried twice in the same plan/apply cycle, **Then** results are stable (no spurious drift caused by unordered list responses; see FR-015).

---

### User Story 4 - Operational cluster lifecycle controls: hibernate and resume (Priority: P3)

A platform engineer wants to hibernate idle clusters to save cost and resume them later, using a Terraform attribute on the cluster resource rather than out-of-band API calls.

**Why this priority**: Hibernate/resume is a cost optimization that is valuable but not on the critical path to a usable provider. Modeling it as a `desired_state` attribute keeps the resource shape clean. Depends on User Story 2 (cluster must exist to be hibernated).

**Independent Test**: Apply a running cluster, set the cluster's `desired_state` attribute to `hibernated`, apply, and verify the cluster transitions to hibernated; flip back to `running`, apply, and verify resume succeeds. Plan-after-apply is clean in both directions.

**Acceptance Scenarios**:

1. **Given** an applied running cluster, **When** the user sets `desired_state = "hibernated"` and applies, **Then** the provider calls `POST /clusters/{id}/hibernate` and the cluster reaches the hibernated state.
2. **Given** a hibernated cluster, **When** the user sets `desired_state = "running"` and applies, **Then** the provider calls `POST /clusters/{id}/resume` and the cluster returns to running.
3. **Given** either transition, **When** the user runs `terraform plan` immediately afterward, **Then** plan reports no changes.

---

### User Story 5 - Non-CRUD operational endpoints exposed as Terraform actions (Priority: P3)

A platform engineer wants to trigger one-shot operational events from Terraform — on a cluster: condemn a host, open / extend / close a support tunnel, scale the Flow Gateway, upgrade the Flow Gateway, start cluster recovery; on a notification: acknowledge it — without modeling these as long-lived managed resources (which would not make sense since they have no persistent state).

**Why this priority**: These are imperative operations that exist in the NC2 API and are required for "all APIs" coverage, but they do not match the managed-resource lifecycle. The cluster-scoped actions depend on User Story 2 (cluster must exist first); the notification-acknowledge action depends on User Story 3 (notifications data source for discovery).

**Independent Test**: For each operational endpoint, invoke the corresponding Terraform action against the prerequisite resource, verify the NC2 API receives the call, and verify the action's reported result reflects the NC2 task outcome.

**Acceptance Scenarios**:

1. **Given** an applied cluster, **When** the user invokes the "condemn host" action with a target host identifier, **Then** the provider calls the condemn-host endpoint and reports task success or failure.
2. **Given** an applied cluster, **When** the user invokes "open support tunnel", "extend support tunnel", or "close support tunnel" actions, **Then** the corresponding NC2 endpoint is called and the result is reported.
3. **Given** an applied cluster, **When** the user invokes "scale Flow Gateway" or "upgrade Flow Gateway" actions, **Then** the corresponding NC2 endpoint is called.
4. **Given** an applied cluster, **When** the user invokes "start cluster recovery", **Then** the corresponding NC2 endpoint is called and a recovery task is started.
5. **Given** an unacknowledged notification discovered via the `nc2_notifications` data source, **When** the user invokes the "acknowledge notification" action, **Then** the provider sends `PATCH /notifications/{id}` with `{"data": {"acknowledged": true}}` and a subsequent read of the data source shows the notification as acknowledged.
6. **Given** any of the above actions, **When** the underlying NC2 task or call fails, **Then** the action surfaces a structured error including the task ID (where applicable), the HTTP status code, and the NC2 error code.

---

### Edge Cases

- An NC2 async task exceeds the configured maximum timeout (default 2 hours per the published reference): the provider must return a timeout error that includes the task ID so the user can investigate in the NC2 console.
- The JWT expires mid-operation (5-minute lifetime): the provider must transparently mint a fresh JWT and retry the in-flight call once, without surfacing the rotation to the user.
- API credentials are expired, missing, or rotated mid-apply: the provider must surface a clear authentication error and a remediation hint (refresh MyNutanix API key, regenerate key ID/issuer) rather than a generic 401.
- A managed resource is deleted out-of-band in NC2 (e.g., via console): the next `terraform plan` must show the resource as needing re-creation, not error out, and not leave stale state.
- A `terraform destroy` is run on `nc2_cloud_account` or `nc2_cloud_account_region` (no API delete exists): the provider must remove the resource from state with a documented plan-time warning, never silently failing and never claiming the cloud-side resource is gone (see FR-006, FR-007).
- An update is attempted on a field that NC2 does not support updating: the provider must mark that field as requiring replacement, so plan shows a destroy + recreate rather than silently ignoring or failing at apply.
- A user sets `access_policy` on `nc2_azure_cluster` or `nc2_gcp_cluster`: validation must fail at plan time, since the access-policy endpoint is AWS-only (see FR-010a).
- Two operations targeting the same cluster overlap (e.g., user issues `apply` while a hibernate is still in flight): the provider must either serialize cleanly or return a clear "operation in progress" error with the conflicting task ID.
- An NC2 endpoint returns a partial response or an unexpected field: the provider must not crash and must preserve all unknown fields on read so that subsequent plans remain drift-free.
- A list-style data source returns results in non-deterministic order: the provider must normalize ordering so that the same query produces the same Terraform output across runs.
- Hibernate is requested on a cluster type or state that does not support it: the provider must reject at plan time where possible, or surface a clear apply-time error otherwise.
- A managed resource is imported via `terraform import`: the resulting state must produce a drift-free plan immediately (no spurious diffs).
- Sensitive fields (API keys, credentials, JWT secrets) appear in logs, plan output, or state: this must never happen — these must always be redacted.
- A new operation is added to `openapi/openapi.json` upstream: the coverage report (FR-021a) must flag it as unmapped so the provider does not silently fall behind.

## Requirements *(mandatory)*

### Functional Requirements

#### Authentication and Provider Configuration

- **FR-001**: The provider MUST authenticate against the NC2 v2 API using a MyNutanix API key, a key ID, and an issuer UUID; mint a short-lived JWT (HS512, `kid` header equal to the key ID, `aud = https://apikeys.nutanix.com`, recommended `exp` of 5 minutes) for each request; and refresh the JWT before expiry without user intervention.
- **FR-001a**: The provider MUST resolve each credential attribute (`api_key`, `key_id`, `issuer`) from up to three sources, in this strict precedence (highest → lowest):
  1. The provider configuration block.
  2. The environment variables `NC2_API_KEY`, `NC2_KEY_ID`, `NC2_ISSUER`.
  3. An optional `credentials_file` (default path `~/.nc2/credentials`, INI or TOML format, with profile selection via a `profile` provider attribute or `NC2_PROFILE` env var).
  When a credential attribute is supplied by more than one source AND the values differ, the provider MUST emit a non-fatal warning naming both sources and the attribute name, and MUST use the higher-precedence value. When a credential attribute is supplied by more than one source AND the values are identical, no warning is emitted.
- **FR-001b**: The provider MUST NOT expose, depend on, or import any native client for external secret stores (HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, 1Password, etc.). Users wishing to source credentials from such stores MUST do so by composing the corresponding Terraform data source into the provider configuration block. This rule MUST be enforced by an import-graph check in CI.
- **FR-002**: The provider MUST treat the following as sensitive — they MUST NOT appear in plan output, apply output, logs, persisted Terraform state, or any debug dump in plaintext, and MUST always be redacted (e.g., to `(sensitive)`):
  - The MyNutanix authentication inputs and derived material: API key, key ID, issuer, derived JWT, and any bytes used to sign the JWT.
  - Every managed-resource or data-source attribute whose JSON field name matches (case-insensitive substring) any of: `credential`, `password`, `secret`, `token`, `private_key`, `api_key`.
  - Every additional attribute explicitly listed in the per-resource "sensitive registry" maintained alongside the provider source code (one registry entry per resource / data source) — used to capture sensitive fields whose names do not match the pattern list.
- **FR-002a**: When a new attribute appears in `openapi/openapi.json` whose JSON field name matches one of the FR-002 patterns, OR which is reachable from an already-sensitive parent attribute, the OpenAPI-coverage build step (FR-021a) MUST emit a lint warning that fails CI in `strict` mode, instructing the maintainer to either confirm the auto-marked sensitive classification or override it in the per-resource sensitive registry. Unknown new fields not matching any pattern default to non-sensitive but also emit a lower-severity classification-needed warning.
- **FR-002b**: The provider MUST ship a unit test per managed resource and per data source asserting that, for every attribute classified as sensitive by FR-002, no plaintext value appears in (a) the rendered plan, (b) the rendered apply output, (c) the persisted state file representation, and (d) the provider's debug log output at the highest verbosity level supported by the framework.
- **FR-003**: The provider MUST allow configurable timeouts for NC2 async task polling, with defaults matching the published NC2 reference (polling interval 60 s, maximum 2 hours). The polling target is `GET /tasks/{id}`; a task is terminal when its status is `done`, `failed`, or `cancelled`. The provider MUST treat `202 Accepted` responses on any create / mutate endpoint as an async signal and switch to polling.
- **FR-003a**: The provider MUST expose the NC2 API base URL as a configurable provider attribute, defaulting to `https://cloud.nutanix.com/api/v2`, so that air-gapped or proxied environments can override it.
- **FR-003b**: The provider MUST enforce TLS certificate verification on every outbound HTTPS call without exception. There MUST be no `insecure_skip_verify` (or equivalent) provider attribute, environment variable, build tag, or hidden flag that disables verification — neither in release nor in development builds. CA trust customization is supported only by:
  1. The operating system's trust store (default), and
  2. An optional `ca_bundle` provider attribute pointing to a PEM-encoded file of additional trusted root certificates, which the provider appends to (not replaces) the OS trust store at startup.
  If `ca_bundle` is set but unreadable or not valid PEM, the provider MUST fail validation at plan time with an actionable error (path, parse offset where applicable).
- **FR-003c**: The provider MUST ship a unit test asserting that no code path constructs an HTTP client with `InsecureSkipVerify: true` or otherwise disables certificate verification, and an acceptance test that confirms operations against a server presenting an untrusted certificate fail with a TLS verification error (not a generic network error).

#### Managed Resources

- **FR-004**: The provider MUST expose NC2 clusters on AWS, Azure, and Google Cloud as three dedicated managed resource types (`nc2_aws_cluster`, `nc2_azure_cluster`, `nc2_gcp_cluster`) — one per cloud, mirroring the `POST /clusters/{aws|azure|gcp}` create endpoints — with full create / read / update / delete lifecycle. Common attributes (name, organization, cloud account, region, capacity, tags, SSH key, license, redundancy, network mode) MUST share consistent naming across the three; cloud-specific attributes MUST live only on the corresponding resource.
- **FR-005**: The provider MUST expose NC2 organizations as managed resources, with create / read / update / terminate lifecycle. Terminate maps to `PATCH /organizations/{id}/terminate`, not to `DELETE`, since the NC2 API does not expose a DELETE on organizations.
- **FR-006**: The provider MUST expose NC2 cloud accounts as managed resources, with create / read / update lifecycle including credential rotation via `POST /cloud-accounts/{id}/update-credentials`. **Caveat — no delete endpoint**: the NC2 v2 API exposes no `DELETE` for cloud accounts. The provider MUST handle `terraform destroy` for `nc2_cloud_account` by (a) emitting a clear warning at plan time that destruction will only remove the resource from Terraform state and that the cloud account must be removed through the NC2 console, and (b) removing the resource from state on apply without calling any NC2 mutation endpoint. The destroy behavior MUST be documented in the resource reference page.
- **FR-007**: The provider MUST expose NC2 cloud account regions as managed resources (`nc2_cloud_account_region`), where adding a region maps to `POST /cloud-accounts/{cloud_account_id}/regions`. **Caveat — no remove endpoint**: the NC2 v2 API exposes no remove-region endpoint; `terraform destroy` MUST follow the same state-only-removal pattern as FR-006, with a plan-time warning and resource-reference documentation.
- **FR-008**: For every managed resource, the create operation MUST handle the NC2 async task model: when an endpoint returns `202 Accepted` with a `data.id` task identifier, the provider MUST poll `GET /tasks/{id}` per FR-003 until terminal, then call the corresponding `GET` to read final state.
- **FR-009**: For every managed resource, the read operation MUST treat NC2 `404 Not Found` responses as state removal (so plan re-creates the resource), not as errors.
- **FR-010**: For every managed resource, the update operation MUST route each changed attribute to the correct NC2 endpoint. For clusters specifically: `capacity` → `POST /clusters/{id}/update-capacity`; `host_access_ssh_key` (or equivalent) → `POST /clusters/{id}/update-ssh-key`; `license` → `POST /clusters/{id}/update-license`; `resource_tags` → `POST /clusters/{id}/update-resource-tags`; `access_policy` (AWS only — see FR-010a) → `POST /clusters/{id}/update-access-policy`; remaining mutable cluster fields → `PATCH /clusters/{id}`.
- **FR-010a**: The `access_policy` attribute and its update endpoint apply **only** to `nc2_aws_cluster`. The Azure and GCP cluster resources MUST NOT expose this attribute; if a user attempts to set it on those resources, validation MUST fail at plan time.
- **FR-010b**: For each NC2 resource that exposes both `PATCH` and `PUT` update variants (clusters, organizations, cloud accounts, notifications), the provider MUST canonically use `PATCH` for partial-field updates driven by Terraform diffs, and reserve `PUT` only for full-replacement scenarios (none expected for normal Terraform flows). This canonical mapping MUST be documented and used consistently.
- **FR-011**: For every managed resource attribute that NC2 does not support updating in place, the provider MUST mark the attribute so that a change forces resource replacement.
- **FR-012**: The provider MUST support cluster hibernate and resume as a `desired_state` attribute on each cluster resource (values: `running`, `hibernated`), mapping transitions to `POST /clusters/{id}/hibernate` and `POST /clusters/{id}/resume`.
- **FR-013**: Every managed resource MUST support `terraform import` such that, after import, `terraform plan` reports no changes.

#### Data Sources

- **FR-014**: The provider MUST expose the following data sources, each backed by the corresponding read endpoint in `openapi/openapi.json`:
  - `nc2_organizations` (list) and `nc2_organization` (single) — `GET /organizations`, `GET /organizations/{id}`
  - `nc2_organization_audit_trail` — `GET /organizations/{id}/audit-trails`
  - `nc2_cloud_accounts` (list, scoped to organization) and `nc2_cloud_account` (single) — `GET /organizations/{id}/cloud-accounts`, `GET /cloud-accounts/{id}`
  - `nc2_cloud_account_regions` — `GET /cloud-accounts/{cloud_account_id}/regions`
  - `nc2_availability_zones` — `GET /cloud-accounts/{cloud_account_id}/regions/{region_id}/availability-zones`
  - `nc2_ssh_keys` — `GET /cloud-accounts/{cloud_account_id}/regions/{region_id}/ssh-keys`
  - `nc2_prism_centrals` — `GET /cloud-accounts/{cloud_account_id}/regions/{region_id}/prism-centrals`
  - `nc2_vnets` — `GET /cloud-accounts/{cloud_account_id}/regions/{region_id}/vnets`
  - `nc2_vpcs` — `GET /cloud-accounts/{cloud_account_id}/regions/{region_id}/vpcs`
  - `nc2_remote_storage_profiles` — `GET /remote-storage-profiles`
  - `nc2_clusters` (list) and `nc2_cluster` (single) — `GET /clusters`, `GET /clusters/{id}`
  - `nc2_cluster_cloud_resources` (cluster-scoped, not account-scoped) — `GET /clusters/{cluster_id}/cloud-resources`
  - `nc2_notifications` (list only — notifications are not user-creatable) — `GET /notifications`
  - `nc2_tasks` (list) and `nc2_task` (single, useful for tracking asynchronous operations referenced by other resources) — `GET /tasks`, `GET /tasks/{id}`
- **FR-014a**: SSH keys, VPCs, VNets, Prism Centrals, availability zones, remote storage profiles, and cloud resources are read-only in the NC2 API and MUST NOT be exposed as managed resources — only as data sources.
- **FR-015**: Data sources that return collections MUST normalize ordering deterministically (e.g., by ID) so that repeated reads produce identical Terraform output across runs.

#### Actions (Operational Endpoints)

- **FR-016**: The provider MUST expose the NC2 endpoints that are imperative one-shots without persistent Terraform-managed state as Terraform actions — one action per endpoint:
  - `nc2_cluster_condemn_host` — `POST /clusters/{id}/condemn-host`
  - `nc2_cluster_open_support_tunnel` — `POST /clusters/{id}/open-support-tunnel`
  - `nc2_cluster_extend_support_tunnel` — `POST /clusters/{id}/extend-support-tunnel`
  - `nc2_cluster_close_support_tunnel` — `POST /clusters/{id}/close-support-tunnel`
  - `nc2_cluster_scale_flow_gateway` — `POST /clusters/{id}/scale-out-fgw`
  - `nc2_cluster_upgrade_flow_gateway` — `POST /clusters/{id}/upgrade-fgw`
  - `nc2_cluster_start_recovery` — `POST /clusters/{id}/start-recovery`
  - `nc2_notification_acknowledge` — `PATCH /notifications/{id}` with `{"data": {"acknowledged": true}}`
- **FR-016a**: The cluster update sub-endpoints (`update-capacity`, `update-ssh-key`, `update-license`, `update-resource-tags`, `update-access-policy`) MUST NOT be exposed as standalone Terraform actions — they are reached only through `nc2_aws_cluster` / `nc2_azure_cluster` / `nc2_gcp_cluster` managed-resource updates per FR-010. Exposing both forms would create two ways to mutate the same state and break drift detection.
- **FR-017**: Each action MUST surface NC2 task results — success, failure, NC2 error code, error message, task identifier — as structured outputs that the user can reference downstream.

#### Quality, Drift, and Destroy Guarantees

- **FR-018**: For every managed resource, immediately running `terraform plan` after `terraform apply` MUST report "No changes" — i.e., the provider MUST be drift-free for every supported configuration.
- **FR-019**: For every managed resource, `terraform destroy` MUST remove the resource from NC2 and from Terraform state, and a subsequent `terraform plan` MUST report no resources to manage and no changes.
- **FR-020**: The provider MUST surface every NC2 API error to the user with at least the HTTP status code, the NC2 error code (if present in the response), the originating endpoint, and an actionable remediation hint where the error is well-known. Well-known statuses to handle explicitly are those listed in the NC2 reference: `200`, `201`, `202`, `400`, `401`, `403`, `404`, `406`, `500`.

#### Observability and Audit

- **FR-020a**: The provider MUST emit exactly one structured audit record per outbound NC2 HTTP call via Terraform's `tflog` facility at INFO level. Each record MUST be a single JSON object containing at least the following fields:
  - `ts` — RFC 3339 timestamp at the moment the call was initiated.
  - `correlation_id` — a UUID v4 generated by the provider at the start of each Terraform operation and propagated across every NC2 call made by that operation.
  - `terraform_op` — the Terraform-side trigger, e.g. `nc2_aws_cluster.create`, `nc2_cloud_account.update`, `nc2_cluster_condemn_host.invoke`, `data.nc2_regions.read`.
  - `http_method` — `GET`, `POST`, `PATCH`, `PUT`.
  - `path` — the NC2 endpoint path (e.g., `/clusters/{id}/hibernate`), parameterized form with identifiers preserved literally.
  - `status` — HTTP status code returned by NC2.
  - `latency_ms` — wall-clock duration of the call in milliseconds.
  - `nc2_task_id` — the task identifier when NC2 returns `202 Accepted` with a task; omitted otherwise.
  - `nc2_error_code` — the NC2 error code when present in an error response; omitted otherwise.
- **FR-020b**: Every audit record MUST redact sensitive fields per FR-002 before serialization. Request bodies and response bodies MUST NOT appear in the audit record; only the structured fields listed in FR-020a do.
- **FR-020c**: The provider MUST NOT expose a private audit-log sink, file path, syslog target, or external transport for these records. Routing and persistence are delegated entirely to Terraform's standard `TF_LOG` and `TF_LOG_PATH` environment variables, so that operators capture audit logs the same way they capture all other Terraform provider logs.
- **FR-020d**: A unit test per managed resource, per data source, and per action MUST assert that performing the operation emits exactly one audit record per NC2 HTTP call made, that all required fields from FR-020a are present, and that no value matching any sensitive-field pattern from FR-002 appears in the rendered JSON.

#### Coverage

- **FR-021**: The provider MUST cover 100% of the operations defined in `openapi/openapi.json` — currently **49 operations across 40 paths**, grouped into 13 tags (`availability zones`, `clusters`, `cloud accounts`, `cloud resources`, `notifications`, `organizations`, `prism centrals`, `regions`, `remote storage profiles`, `ssh keys`, `tasks`, `vnets`, `vpcs`). Every operation MUST be reachable through either a managed resource, a data source, or an action. A coverage report MUST be generated from the OpenAPI document (not from documentation pages) and MUST run in CI; any operation present in `openapi/openapi.json` but not mapped MUST fail the build.
- **FR-021a**: When `openapi/openapi.json` is updated (new endpoints, new fields, new error responses), the coverage report MUST flag any new operation as unmapped on the next CI run, so that the provider does not silently fall behind the API.

#### Testing

- **FR-022**: Every managed resource and every data source MUST ship with unit tests covering: schema definition, attribute validation, plan modifiers, and each CRUD/read method. Unit tests MUST NOT require live NC2 access and MUST run on every pull request.
- **FR-023**: Every managed resource MUST ship with an acceptance test ("E2E") that performs the full lifecycle: `terraform init` → `terraform plan` (shows create) → `terraform apply` → `terraform plan` (verifies "No changes") → for resources that support in-place updates, an update step followed by another zero-diff plan → `terraform destroy` → `terraform plan` (verifies clean empty state).
- **FR-024**: Every data source MUST ship with an acceptance test that exercises the data source against live NC2 and asserts on returned shape and content.
- **FR-025**: Every action MUST ship with an acceptance test that invokes the action against a prerequisite cluster and asserts that the NC2 task completes with the expected outcome.
- **FR-026**: Acceptance tests MUST execute against a real NC2 environment and MUST be runnable from a single command (e.g., `make testacc`). They MAY be gated by an environment flag to avoid accidental billable runs.
- **FR-027**: The test suite MUST fail the build if any new public function, managed resource, or data source is added without corresponding tests (test-required gate).

#### Documentation

- **FR-028**: Every exported Go symbol (package, type, function, method, constant) in the provider source code MUST carry an inline doc comment describing purpose, parameters, returned values, and error conditions where applicable.
- **FR-029**: The provider MUST ship general documentation that includes, at minimum: a provider overview, a configuration guide, an authentication guide, a getting-started tutorial, one reference page per managed resource, one reference page per data source, one reference page per action, and a runnable example per managed resource.
- **FR-030**: Documentation pages MUST be generated from the provider schema where the framework supports schema-driven generation, so that documentation cannot drift from the implemented schema.
- **FR-031**: The provider MUST ship a changelog (Keep-a-Changelog format) and a migration guide entry for any backwards-incompatible change.

#### Release and Distribution

- **FR-032**: The provider MUST be releasable through the Terraform Registry, with a versioned `terraform-registry-manifest.json`, signed release artifacts, and a documented release process. The signing posture MUST layer the following, all generated automatically by the release pipeline:
  1. **GPG signature** over the release artifacts and the `SHA256SUMS` file, using a key whose public half is registered with the Terraform Registry (Terraform Registry compatibility requirement).
  2. **Sigstore cosign keyless signatures** for each release artifact (one `.sig` and one `.pem` per artifact), published alongside the GPG-signed bundle.
  3. **SLSA Build Level 3 provenance attestation** (e.g., generated via `slsa-github-generator` or equivalent), published as an in-toto attestation alongside the release. The attestation MUST identify the source repository commit, the builder identity, and the build parameters.
- **FR-032a**: The release pipeline MUST run `govulncheck` against the compiled binary AND `osv-scanner` against `go.mod` / `go.sum` on every release. The pipeline MUST fail (and the release MUST NOT publish) if either tool reports a vulnerability of severity HIGH or CRITICAL in any direct or transitive dependency. Vulnerabilities of severity MEDIUM or LOW MUST be recorded in the release notes but do not block the release.
- **FR-032b**: The release pipeline MUST publish, alongside each release, a signed `provenance.intoto.jsonl` (SLSA), a `SHA256SUMS` file, a `SHA256SUMS.sig` (GPG), per-artifact Sigstore `.sig` and `.pem` files, and a structured `vulnerability-report.json` summarizing the `govulncheck` and `osv-scanner` results that were evaluated for the gate.

### Key Entities *(include if feature involves data)*

- **Organization**: An NC2 tenant; root of the resource hierarchy. Owns cloud accounts and is the unit of billing and audit.
- **Cloud Account**: A credentialed account on AWS, Azure, or GCP that NC2 uses to provision infrastructure on behalf of the organization. Owns regions, SSH keys, VPCs/VNets, and clusters.
- **Region**: A cloud provider region enabled on a cloud account; gates which AZs and clusters are available.
- **Availability Zone**: A failure domain within a region; cluster placement target.
- **Cluster**: A Nutanix-managed cluster running on a specific cloud account in a specific region/AZ. Has a type (AWS, Azure, GCP) that determines its cloud-specific attribute surface. Supports lifecycle states including provisioning, running, hibernated, resuming, terminating. Owns capacity, SSH key, license, tags, access policy, and Flow Gateway configuration.
- **VPC / VNet**: Cloud-provider network constructs visible to the cloud account, referenced by cluster network configuration.
- **SSH Key**: A key registered on the cloud account, attachable to clusters.
- **Prism Central**: A Nutanix management plane endpoint available to a cluster.
- **Remote Storage Profile**: A storage configuration profile attachable to a cluster.
- **Cloud Resource**: A generic cloud-side artifact (read-only inventory).
- **Notification**: An NC2 platform notification with an acknowledged/unacknowledged state.
- **Task**: An asynchronous NC2 operation tracking record with status, progress, and result; the polling target for every async API call.
- **Action Invocation** (not persisted): A one-shot operational call — condemn host, open / close / extend support tunnel, scale Flow Gateway, upgrade Flow Gateway, start cluster recovery, acknowledge notification — that produces a task but no managed-resource state.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of the operations defined in `openapi/openapi.json` (currently 49 operations across 40 paths) are reachable from the provider as a managed resource, data source, or action, verified by an automated coverage report generated from the OpenAPI document.
- **SC-002**: Across every managed resource type, 100% of acceptance test runs report "No changes" when `terraform plan` is executed immediately after `terraform apply` (drift-free guarantee).
- **SC-003**: Across every managed resource type, 100% of acceptance test runs show a clean `terraform plan` (no resources, no changes) after `terraform destroy` (clean-destroy guarantee).
- **SC-004**: A user starting from an empty workspace and an empty NC2 organization can, following the documented examples, (a) provision an organization and a cloud account, then (b) provision a working NC2 cluster on any of the three supported clouds (AWS, Azure, GCP), in under 120 minutes elapsed end-to-end, of which no more than 5 minutes is provider-induced overhead (the remainder being NC2 provisioning latency).
- **SC-005**: Unit test line coverage on the provider source is at least 80% at first release and is enforced by CI.
- **SC-006**: 100% of managed resources, data sources, and actions have a published documentation page with at least one runnable example.
- **SC-007**: 100% of exported Go symbols in the provider source have inline doc comments, enforced by a linter in CI.
- **SC-008**: When the NC2 API returns an error, the user-visible Terraform error message includes the HTTP status, the NC2 error code, and an actionable next step in at least 90% of distinct error scenarios catalogued in tests.
- **SC-009**: Every managed resource is importable via `terraform import` such that the post-import plan reports no changes; verified by an acceptance test per resource.
- **SC-010**: First public release is consumable from the Terraform Registry with signed artifacts and a versioned manifest.

## Assumptions

- The authoritative API surface is the local file `openapi/openapi.json` (OpenAPI 3.0.0, NC2 API version `v2`, server `https://cloud.nutanix.com/api/v2`). The public reference at `https://www.nutanix.dev/api_reference/apis/nc2.html` is treated as a mirror, not a source of truth; if the two diverge, the local OpenAPI wins.
- The implementation will use the latest generally-available Terraform Provider Framework (the `terraform-plugin-framework` line, currently 1.x), including its support for actions and ephemeral resources where they best fit non-CRUD operations. This is treated as a project-level technology decision rather than an open question, per the user's request for "the latest terraform provider framework."
- The implementation language is Go (the only language Terraform plugins are written in); the toolchain target is the most recent Go release supported by the framework.
- All three NC2 cloud backends (AWS, Azure, GCP) are in scope, since the user explicitly requested coverage of "all API for Nutanix NC2 as documented."
- Authentication follows the MyNutanix API key + key ID + issuer UUID + JWT (HS512) flow described in the OpenAPI's `info.description`: the JWT has `aud = https://apikeys.nutanix.com`, a `kid` header equal to the key ID, and a 5-minute recommended lifetime. The provider refreshes the JWT automatically.
- Long-running endpoints (`POST /clusters/{aws|azure|gcp}`, `terminate`, `hibernate`, `resume`, and the cluster operational endpoints) return `202 Accepted` with a task object in `data.id`; the polling target `GET /tasks/{id}` reports terminal status as `done`, `failed`, or `cancelled`. The provider implements this exact contract.
- The NC2 v2 API exposes **no DELETE for cloud accounts** and **no remove-region endpoint**. `terraform destroy` on `nc2_cloud_account` and `nc2_cloud_account_region` therefore performs state-only removal with a plan-time warning (see FR-006, FR-007). If Nutanix later publishes delete endpoints in `openapi/openapi.json`, the destroy behavior will be upgraded to call them.
- Notifications are system-generated (no user-create endpoint exists); they are exposed as a data source plus an acknowledge action, not as a managed resource.
- `POST /clusters/{id}/update-access-policy` is AWS-only per the OpenAPI summary; the attribute and its update path appear only on `nc2_aws_cluster`.
- For NC2 endpoints that expose both `PATCH` and `PUT` update variants (clusters, organizations, cloud accounts, notifications), the provider canonically uses `PATCH` for Terraform diff-driven updates and reserves `PUT` for an explicit full-replacement code path that normal Terraform flows do not exercise.
- The release channel for the provider is the public Terraform Registry (under either the Nutanix namespace or a community namespace, to be decided at release time). The release process is GoReleaser + signed artifacts, which is the Terraform Registry convention.
- Acceptance ("E2E") tests require a live NC2 account and incur cloud spend; they run on demand via a `make testacc` target gated by `TF_ACC=1` and credentials, and on a scheduled CI lane — not on every pull request. Unit tests run on every pull request.
- Non-CRUD operational endpoints (condemn host, support tunnel control, Flow Gateway scale/upgrade, cluster recovery, acknowledge notification) are modeled as framework "actions" rather than managed resources, since they are imperative one-shots with no persistent state. If a future Terraform release deprecates or reshapes the actions concept, the provider will follow upstream guidance.
- Cluster hibernate and resume are modeled as a single `desired_state` attribute on the cluster resource (values: `running`, `hibernated`) rather than as two separate action resources, because they are state transitions of an existing managed resource.
- "Drift-free plan" and "clean destroy" are quality bars that apply to every managed resource without exception; they are not negotiable per-resource.
- Inline documentation is enforced via a Go documentation linter (e.g., `revive` with the `exported` rule or equivalent) in CI; general documentation is enforced by checking that every resource/data-source/action has a corresponding docs file before release.
- API coverage is enforced by a tool that parses `openapi/openapi.json` and cross-checks every operation against the provider's resource / data-source / action registry; missing mappings fail CI.
