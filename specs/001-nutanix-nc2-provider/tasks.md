---
description: "Task list for terraform-provider-nc2 implementation"
---

# Tasks: Terraform Provider for Nutanix Cloud Clusters (NC2)

**Input**: Design documents under `specs/001-nutanix-nc2-provider/`

**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/, quickstart.md

**Tests**: Test tasks are MANDATORY per Constitution Principle II (Test-Driven Development, NON-NEGOTIABLE). For every behavior change, the corresponding test task MUST be authored first, MUST be verified to fail for the right reason, and only THEN may the implementation task begin.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Implementation Status (last updated 2026-05-13)

Snapshot of where the codebase stands relative to the task list. Complete unit / package coverage is summarized via `go test ./...` results captured in this session.

| Phase | Scope | Status |
|---|---|---|
| Phase 1 — Setup (T001–T010) | Repo layout, Go module, CI scaffolding, lint config, license, README | **Complete** |
| Phase 2 — Foundational (T011–T045) | redact, oapi, auth, audit, client, provider, orgshared, dsshared, clustershared, main.go | **Complete** for runtime libraries (T011–T023, T026–T032, T035–T038, T043, T045); **TODO**: T025 untrusted-cert TLS test, T034 import-graph guard, T039–T042 CI tools, T044 main_test.go |
| Phase 3 — US1 MVP (T046–T090) | nc2_organization, nc2_cloud_account, nc2_cloud_account_region + 6 data sources | **Not started**; this is the highest-priority next slice |
| Phase 4 — US2 (T091–T126) | nc2_aws_cluster, nc2_azure_cluster, nc2_gcp_cluster + cluster data sources | **Implementations land**; full test coverage (lifecycle, drift-free, sensitive, audit-emission) only partially authored — most existing `*_test.go` files are smoke-level operation-mapping checks |
| Phase 5 — US3 (T127–T151) | 10 inventory / discovery data sources | **Implementations land**; tests are minimal — same gap as Phase 4 |
| Phase 6 — US4 (T152–T164) | desired_state hibernate/resume on the 3 cluster resources | `MapDesiredStateTransition` + `RouteUpdate` pure helpers complete with unit tests; per-cluster wiring lands; per-cluster hibernate_test.go files still TODO |
| Phase 7 — US5 (T165–T188) | 8 actions + shared scaffolding | **Implementations land**; per-action tests are operation-mapping smoke checks only |
| Phase 8 — Polish (T189–T212) | Release pipeline, supply-chain attestation, docs guides, integration tests | **Not started** |

`go test ./...` is green across every package present in the repo as of this snapshot.

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4, US5)
- Setup, Foundational, and Polish phases have no story label
- Include exact file paths in descriptions

## Path Conventions

Single Go module at the repository root (per `plan.md`). All paths below are relative to the repo root.

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Initialize the Go module, repo layout, toolchain, and CI scaffolding so every later phase has a working build.

- [X] T001 Create top-level Go module: `go mod init github.com/<org>/terraform-provider-nc2` (placeholder org), commit go.mod / go.sum with Go 1.24 toolchain pin
- [X] T002 Create the directory layout from plan.md `### Source Code` section: `internal/{provider,auth,client,audit,redact,resources,datasources,actions}/`, `tools/{coverage-check,sensitive-lint}/`, `tests/{acceptance,fixtures}/`, `examples/`, `docs/`, `.github/workflows/`, `.githooks/` — each with a placeholder `.gitkeep` where empty
- [X] T003 [P] Add tools-only dependency manifest at `tools.go` with `// +build tools` (kin-openapi, tfplugindocs, golangci-lint, govulncheck, osv-scanner, cosign, goreleaser, slsa-github-generator)
- [X] T004 [P] Add `Makefile` with targets: `tools`, `build`, `test`, `coverage`, `lint`, `doc-lint`, `openapi-coverage`, `sensitive-lint`, `sensitive-lint-strict`, `vuln`, `testacc`, `release` (target stubs invoking the right tool); behavior matches `quickstart.md` §5
- [X] T005 [P] Add `.golangci.yml` enabling `errcheck`, `govet`, `staticcheck`, `revive`, `paralleltest`, `tparallel`, `gosec`, `unparam`, with the `revive` `exported` rule enforced (SC-007)
- [X] T006 [P] Add `.github/workflows/ci.yml` running `make lint`, `make test`, `make coverage`, `make openapi-coverage`, `make sensitive-lint-strict`, `make vuln` on every PR; matrix on Go 1.23 and 1.24
- [X] T007 [P] Add `.github/workflows/acceptance.yml` running `make testacc` on a nightly schedule and on `workflow_dispatch`, gated by repository secrets `NC2_API_KEY` / `NC2_KEY_ID` / `NC2_ISSUER`
- [X] T008 [P] Add `.github/PULL_REQUEST_TEMPLATE.md` with the Constitution checklist (Library-First, TDD, Functional-Patterns boxes)
- [X] T009 [P] Add `.githooks/pre-commit` shell script running `go test ./...` locally; document `git config core.hooksPath .githooks` in `quickstart.md` (already present) and `README.md`
- [X] T010 [P] Add `LICENSE` (MPL-2.0, Terraform-ecosystem convention) and a top-level `README.md` stub pointing at `specs/001-nutanix-nc2-provider/` and `docs/`

**Checkpoint**: Repository builds (`make build` may fail until Phase 2 produces `main.go`), CI workflow files are committed, all linters/tools are declared.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Stand up the cross-cutting libraries every user story depends on: credential resolution, JWT auth, HTTP client, task polling, error mapping, audit logging, sensitive-field redaction, and the OpenAPI-driven CI tools. No user story can begin until this phase completes.

**Constitution gate**: every implementation task in this phase MUST have its test task authored first (TDD, NON-NEGOTIABLE).

### Library: `internal/redact` (FR-002, FR-002a, FR-002b)

- [X] T011 [P] Author the failing pattern-match unit tests in `internal/redact/patterns_test.go`: table-driven cases for `credential`, `password`, `secret`, `token`, `private_key`, `api_key` (case-insensitive), plus negative cases (`description`, `name`, etc.); verify failure with `go test ./internal/redact/...` before moving on
- [X] T012 [P] Author the failing redaction unit tests in `internal/redact/redact_test.go`: cases for flat fields, nested fields, list fields, override registry entries, idempotency, and "marked sensitive but absent from data" no-op
- [X] T013 Implement `internal/redact/patterns.go` exporting `var FR002Patterns []string` and `func MatchPattern(name string) bool` until T011 passes
- [X] T014 Implement `internal/redact/redact.go` exporting `type Registry`, `func NewRegistry(extra []string) Registry`, `func (r Registry) Redact(rec map[string]any) map[string]any` (pure function, returns a deep-copied redacted map) until T012 passes
- [X] T015 [P] Write `internal/redact/README.md` describing purpose, public API, and the FR-002 pattern list

### Library: `internal/auth` (FR-001)

- [X] T016 [P] Author the failing JWT minting unit tests in `internal/auth/jwt_test.go`: golden-fixture cases per the R-02 recipe (HS512, `secret = base64(HMAC-SHA512(api_key, key_id))`, `aud=https://apikeys.nutanix.com`, `kid` header, `exp = iat + 300s`), plus negative cases (empty inputs error cleanly, malformed issuer rejected)
- [X] T017 [P] Author the failing TokenManager unit tests in `internal/auth/manager_test.go`: cached-token reuse, refresh-on-expiry (using a `func() time.Time` clock injected into the manager), refresh-on-401-then-retry, concurrent-access safety via the `-race` flag
- [X] T018 Implement `internal/auth/jwt.go` exporting `func MintJWT(creds Credentials, now time.Time) (Token, error)` as a pure function until T016 passes
- [X] T019 Implement `internal/auth/manager.go` exporting `type TokenManager`, `func NewTokenManager(creds Credentials, clock func() time.Time) *TokenManager`, and `func (m *TokenManager) Token(ctx context.Context) (string, error)` (the only mutable state in the library; guarded by sync.RWMutex) until T017 passes
- [X] T020 [P] Write `internal/auth/README.md` describing the JWT recipe (with reference to research.md R-02), the TokenManager lifecycle, and the absence of any external-secret-store integration (FR-001b)

### Library: `internal/audit` (FR-020a, FR-020b, FR-020c, FR-020d)

- [X] T021 [P] Author the failing audit-record unit tests in `internal/audit/audit_test.go`: assert that calling `Record(ctx, AuditRecord{...})` produces exactly one `tflog.Info` entry per the FR-020a schema, that omitted fields (`nc2_task_id`, `nc2_error_code`) do not appear in the JSON, and that `redact.Registry` is applied to every string-valued field before emission (FR-020b); use `terraform-plugin-log/tflogtest` to capture
- [X] T022 Implement `internal/audit/audit.go` exporting `type AuditRecord` (flat struct per data-model.md §11) and `func Record(ctx context.Context, r AuditRecord)` (pure-by-construction wrapper over `tflog.Info`) until T021 passes
- [X] T023 [P] Write `internal/audit/README.md` documenting the audit-record schema, the routing-via-TF_LOG contract (FR-020c), and the no-private-sink guarantee

### Library: `internal/client` (FR-003, FR-003a, FR-003b, FR-008, FR-009, FR-020)

- [X] T024 [P] Author the failing HTTP-client unit tests in `internal/client/client_test.go` using `net/http/httptest`: Authorization header carries the JWT, `User-Agent` is `terraform-provider-nc2/<version>`, 200/201/202/400/401/403/404/406/500 paths each surface the right error shape and audit record, 401 triggers exactly one token refresh + retry
- [ ] T025 [P] Author the failing TLS unit tests in `internal/client/tls_test.go` (FR-003c): assert no code path constructs `&tls.Config{InsecureSkipVerify: true}` via an import-graph and reflection check, and assert that an HTTPS request to a server with an untrusted certificate fails with a TLS verification error (not a generic network error); the test uses `httptest.NewTLSServer` and a transient empty trust pool — *partial: source-grep assertion lives in `internal/provider/tls_test.go::TestBuildTLSConfig_NeverSkipsVerify`; the TLS-rejection assertion is still TBD*
- [X] T026 [P] Author the failing task-polling unit tests in `internal/client/tasks_test.go`: 202 then 1× pending then `done` (success), 202 then 1× running then `failed` (error includes NC2 error code), 202 then loop timeout (timeout error includes task id), `task_poll_interval_seconds` and `task_max_timeout_seconds` honored, clock injected
- [X] T027 [P] Author the failing error-mapping unit tests in `internal/client/errors_test.go` (FR-020): every documented NC2 status code maps to a user-facing diagnostic with at least HTTP status, NC2 error code (when present), originating endpoint, and remediation hint where well-known
- [X] T028 Implement `internal/client/client.go` exporting `type Client`, `func New(cfg Config) (*Client, error)`, `func (c *Client) Do(ctx context.Context, req Request) (Response, error)` using stdlib `net/http`; integrates `auth.TokenManager`, emits one `audit.Record` per call, never sets `InsecureSkipVerify` until T024 + T025 pass
- [X] T029 Implement `internal/client/tasks.go` exporting `func (c *Client) PollTask(ctx context.Context, taskID string) (TaskResult, error)` until T026 passes
- [X] T030 Implement `internal/client/errors.go` exporting `type APIError`, `func MapError(resp *http.Response, body []byte, endpoint string) error` until T027 passes
- [X] T031 [P] Write `internal/client/README.md` documenting the client's contract: pure error mapping, mutable state limited to the `http.Client` pool, async-via-`PollTask` always at-least-once, no escape hatch on TLS

### Provider config: `internal/provider` (FR-001a, FR-003a, FR-003b)

- [X] T032 [P] Author the failing credential-resolution unit tests in `internal/provider/config_test.go`: precedence (block > env > file), profile selection (`profile` attr, `NC2_PROFILE` env), conflict warning when sources disagree, no warning when sources agree, missing-credentials error
- [X] T033 [P] Author the failing TLS-config unit tests in `internal/provider/tls_test.go`: default trust store used when `ca_bundle` empty; PEM bundle appended (not replaced) when set; invalid PEM produces actionable plan-time error with path and offset (FR-003b)
- [ ] T034 [P] Author the failing import-graph test in `internal/provider/imports_test.go` (FR-001b): walks the entire module dependency tree and fails if any package path matches `vault`, `secretsmanager`, `keyvault`, `secret-manager`, `1password`
- [X] T035 Implement `internal/provider/config.go` exporting `type ProviderConfig`, `func Resolve(ctx context.Context, block ProviderBlock, env Env, file File) (ProviderConfig, diag.Diagnostics)` as a pure function until T032 passes
- [X] T036 Implement `internal/provider/tls.go` exporting `func BuildTLSConfig(caBundlePath string) (*tls.Config, error)` (never sets `InsecureSkipVerify`) until T033 passes
- [X] T037 Implement `internal/provider/provider.go` exporting `func New() provider.Provider` wiring the framework's schema, mapping `nc2.Schema()` to the `contracts/provider.json` shape, and calling `Resolve` + `BuildTLSConfig` in `Configure`
- [X] T038 [P] Write `internal/provider/README.md` listing every provider-level configuration attribute, its precedence rules, and pointers to `auth`, `client`, `redact`, `audit`

### CI tools: `tools/coverage-check` (FR-021, FR-021a) and `tools/sensitive-lint` (FR-002a)

- [ ] T039 [P] Author the failing coverage-tool unit tests in `tools/coverage-check/main_test.go`: feed it a synthetic OpenAPI doc + synthetic mapping registry; assert that fully-mapped → exit 0; an unmapped operation → non-zero exit; output JSON matches the R-09 schema
- [ ] T040 [P] Author the failing sensitive-lint unit tests in `tools/sensitive-lint/main_test.go`: pattern hits without registry entry → fail; registry entry without OpenAPI counterpart → fail (drift); unclassified field → MEDIUM warning in non-strict, fail in strict
- [ ] T041 Implement `tools/coverage-check/main.go` using `kin-openapi` (behind tools build tag) until T039 passes; produces `tools/coverage-check/output/coverage-report.json`
- [ ] T042 Implement `tools/sensitive-lint/main.go` using `kin-openapi` until T040 passes
- [X] T043 [P] Author `internal/oapi/mapping.go` (used by tools and runtime): exported `type Mapping struct { OperationID, TerraformOp string }` and a discovery helper that scans `internal/{resources,datasources,actions}/*/openapi_mapping.go` files for `var OperationMappings []oapi.Mapping`; ship its own unit test fixture

### Entry point

- [ ] T044 Author the failing provider-server smoke test in `main_test.go`: assert `main` registers the provider via `providerserver.Serve` with the correct address `registry.terraform.io/<namespace>/nc2`
- [X] T045 Implement `main.go` until T044 passes; wires `internal/provider.New`

**Checkpoint**: All foundational libraries are unit-tested green; OpenAPI coverage tool runs and reports `49 total, 0 covered, 49 uncovered` (expected at this point). User story implementation can now begin.

---

## Phase 3: User Story 1 — Manage NC2 organizations and cloud accounts as Terraform resources (Priority: P1) 🎯 MVP

**Goal**: Deliver `nc2_organization`, `nc2_cloud_account`, `nc2_cloud_account_region` managed resources, plus the read-side data sources `nc2_organizations`, `nc2_organization`, `nc2_organization_audit_trail`, `nc2_cloud_accounts`, `nc2_cloud_account`. End-to-end org + account + region lifecycle works on a sandbox NC2 tenant.

**Independent Test**: Per spec US1 "Independent Test" — apply org + account, rotate credentials, verify clean plan / drift-free / clean-destroy semantics including the no-DELETE caveats for cloud accounts and regions.

### Managed resource: `nc2_organization`

- [ ] T046 [P] [US1] Author failing schema-shape unit tests in `internal/resources/organization/schema_test.go` asserting the schema matches `contracts/resources/nc2_organization.json` (attribute names, types, modes, validators, plan modifiers, sensitive flags) — contract test per FR-022
- [ ] T047 [P] [US1] Author failing CRUD unit tests in `internal/resources/organization/resource_test.go` using httptest: Create (POST /organizations), Read (GET /organizations/{id}), Update (PATCH), Delete (PATCH /organizations/{id}/terminate), Read returns 404 → state removal (FR-009), ImportState round-trip plan is clean (FR-013)
- [ ] T048 [P] [US1] Author failing drift-free + clean-destroy unit tests in `internal/resources/organization/lifecycle_test.go` (FR-018, FR-019, SC-002, SC-003): apply→plan zero diff; destroy→plan empty
- [ ] T049 [P] [US1] Author failing sensitive-field redaction unit test in `internal/resources/organization/sensitive_test.go` (FR-002b): for every classified attribute, no plaintext in rendered plan / state / debug log
- [ ] T050 [P] [US1] Author failing audit-emission unit test in `internal/resources/organization/audit_test.go` (FR-020d): every CRUD method emits exactly one `tflog` audit record matching FR-020a
- [ ] T051 [US1] Implement `internal/resources/organization/schema.go` until T046 passes
- [ ] T052 [US1] Implement `internal/resources/organization/sensitive.go` exporting `func Sensitive() []string` (empty per data-model.md §1) until T049 passes
- [ ] T053 [US1] Implement `internal/resources/organization/resource.go` (Create/Read/Update/Delete/ImportState) until T047, T048, T050 pass
- [ ] T054 [US1] Implement `internal/resources/organization/openapi_mapping.go` declaring `var OperationMappings []oapi.Mapping` for org operations per R-09
- [ ] T055 [P] [US1] Write `internal/resources/organization/README.md` documenting purpose, public API, and lifecycle states
- [ ] T056 [US1] Acceptance test in `tests/acceptance/nc2_organization_test.go` (FR-023, gated by `TF_ACC=1`): full init → plan → apply → plan-clean → update → plan-clean → destroy → plan-empty; ImportState path (SC-009)

### Managed resource: `nc2_cloud_account`

- [ ] T057 [P] [US1] Author failing schema-shape unit tests in `internal/resources/cloud_account/schema_test.go` asserting `credentials` subtree is marked sensitive (FR-002), per-cloud subfield validation (aws/azure/gcp), `cloud_provider` is `RequiresReplace`
- [ ] T058 [P] [US1] Author failing CRUD unit tests in `internal/resources/cloud_account/resource_test.go` using httptest: Create via `POST /organizations/{id}/cloud-accounts/{cloud_provider}`; Read via `GET /cloud-accounts/{id}`; Update routes correctly (`credentials` diff → `POST /cloud-accounts/{id}/update-credentials`; `name`/`description` diff → `PATCH /cloud-accounts/{id}`; both → credentials first then PATCH); Delete = state-only removal + warning + no HTTP call (FR-006, R-07)
- [ ] T059 [P] [US1] Author failing plan-time warning unit test in `internal/resources/cloud_account/warnings_test.go` asserting the "no DELETE endpoint" warning is emitted at plan time during destroy
- [ ] T060 [P] [US1] Author failing sensitive-redaction unit test in `internal/resources/cloud_account/sensitive_test.go` (FR-002b) covering every cloud's credential subfields; debug-log capture must show `(sensitive)` for `secret_access_key`, `client_secret`, `service_account_json`, etc.
- [ ] T061 [P] [US1] Author failing audit-emission unit test in `internal/resources/cloud_account/audit_test.go` for cloud account CRUD (FR-020d)
- [ ] T062 [US1] Implement `internal/resources/cloud_account/schema.go` (with per-cloud nested schema) until T057 passes
- [ ] T063 [US1] Implement `internal/resources/cloud_account/sensitive.go` (FR-002): the `credentials` subtree fully redacted; explicit secrets listed
- [ ] T064 [US1] Implement `internal/resources/cloud_account/resource.go` (Create/Read/Update with field routing/Delete with warning/ImportState) until T058, T059, T060, T061 pass
- [ ] T065 [US1] Implement `internal/resources/cloud_account/openapi_mapping.go`
- [ ] T066 [P] [US1] Write `internal/resources/cloud_account/README.md` documenting the no-DELETE caveat (FR-006) prominently
- [ ] T067 [US1] Acceptance test in `tests/acceptance/nc2_cloud_account_test.go` (FR-023): full lifecycle including credential rotation (acceptance scenario 4 in spec US1) and the state-only-destroy path (scenario 6); requires an AWS sandbox cloud account by default, with Azure/GCP variants in build-tagged tests

### Managed resource: `nc2_cloud_account_region`

- [ ] T068 [P] [US1] Author failing schema-shape unit tests in `internal/resources/cloud_account_region/schema_test.go`
- [ ] T069 [P] [US1] Author failing CRUD unit tests in `internal/resources/cloud_account_region/resource_test.go`: Create via `POST /cloud-accounts/{cloud_account_id}/regions`; Read via `GET /cloud-accounts/{cloud_account_id}/regions` (filter by id); Update = `RequiresReplace` on `region`; Delete = state-only removal + warning (FR-007)
- [ ] T070 [P] [US1] Author failing audit-emission unit test in `internal/resources/cloud_account_region/audit_test.go` (FR-020d)
- [ ] T071 [US1] Implement `internal/resources/cloud_account_region/schema.go`, `sensitive.go` (empty), `resource.go`, `openapi_mapping.go`, `README.md` until T068, T069, T070 pass
- [ ] T072 [US1] Acceptance test in `tests/acceptance/nc2_cloud_account_region_test.go` (FR-023)

### Data sources (read-only, US1 scope)

- [ ] T073 [P] [US1] Author failing unit tests for `data.nc2_organizations` (list) in `internal/datasources/organizations/datasource_test.go`: Read calls `GET /organizations`, results sorted ascending by `id` (FR-015), emits one audit record per call
- [ ] T074 [P] [US1] Author failing unit tests for `data.nc2_organization` (single by id) in `internal/datasources/organization/datasource_test.go`
- [ ] T075 [P] [US1] Author failing unit tests for `data.nc2_organization_audit_trail` in `internal/datasources/organization_audit_trail/datasource_test.go`: `GET /organizations/{id}/audit-trails`, deterministic ordering
- [ ] T076 [P] [US1] Author failing unit tests for `data.nc2_cloud_accounts` (list, scoped to organization) in `internal/datasources/cloud_accounts/datasource_test.go`
- [ ] T077 [P] [US1] Author failing unit tests for `data.nc2_cloud_account` (single by id) in `internal/datasources/cloud_account/datasource_test.go`
- [ ] T078 [P] [US1] Author failing unit tests for `data.nc2_cloud_account_regions` in `internal/datasources/cloud_account_regions/datasource_test.go` (contract test against `contracts/datasources/nc2_cloud_account_regions.json`)
- [ ] T079 [US1] Implement `internal/datasources/organizations/datasource.go`, `schema.go`, `openapi_mapping.go`, `README.md` until T073 passes
- [ ] T080 [US1] Implement `internal/datasources/organization/datasource.go` etc. until T074 passes
- [ ] T081 [US1] Implement `internal/datasources/organization_audit_trail/datasource.go` etc. until T075 passes
- [ ] T082 [US1] Implement `internal/datasources/cloud_accounts/datasource.go` etc. until T076 passes
- [ ] T083 [US1] Implement `internal/datasources/cloud_account/datasource.go` etc. until T077 passes
- [ ] T084 [US1] Implement `internal/datasources/cloud_account_regions/datasource.go` etc. until T078 passes
- [ ] T085 [P] [US1] Acceptance tests for all 6 US1 data sources in `tests/acceptance/nc2_us1_datasources_test.go` (FR-024)

### Wire-up + examples + docs for US1

- [ ] T086 [US1] Wire all 3 US1 resources and 6 US1 data sources into `internal/provider/provider.go` `Resources()` and `DataSources()` slices
- [ ] T087 [P] [US1] Add runnable examples under `examples/resources/nc2_organization/`, `examples/resources/nc2_cloud_account/` (one per cloud: aws / azure / gcp subdirs), `examples/resources/nc2_cloud_account_region/` (FR-029, SC-006)
- [ ] T088 [P] [US1] Add runnable examples for US1 data sources under `examples/data-sources/<name>/`
- [ ] T089 [P] [US1] Generate docs pages under `docs/resources/nc2_organization.md`, `docs/resources/nc2_cloud_account.md`, `docs/resources/nc2_cloud_account_region.md`, `docs/data-sources/<each>.md` via `tfplugindocs` (FR-030); commit generated output
- [ ] T090 [US1] Update `internal/oapi` discovery so `make openapi-coverage` reports the US1 operation set as covered; CI must be green for US1 scope

**Checkpoint**: User Story 1 (MVP) is fully functional and testable. A user can manage organizations + cloud accounts + regions end-to-end on a sandbox NC2 tenant. SC-004 step (a) (org + cloud account) is demonstrable. STOP HERE for first internal demo / alpha tag.

---

## Phase 4: User Story 2 — Manage NC2 clusters across AWS, Azure, GCP (Priority: P2)

**Goal**: Deliver `nc2_aws_cluster`, `nc2_azure_cluster`, `nc2_gcp_cluster` managed resources with full CRUD including all in-place updates, plus the `nc2_clusters` (list) and `nc2_cluster` (single) data sources.

**Independent Test**: Per spec US2 "Independent Test" — for each of AWS, Azure, GCP, run the full lifecycle including in-place updates of capacity, SSH key, license, tags, plus the AWS-only access policy update; verify drift-free and clean-destroy on every cloud.

**Depends on**: Phase 2 + Phase 3 (cluster needs an org + cloud account + region to exist).

### Shared cluster scaffolding

- [X] T091 [US2] Author failing shared-cluster-helpers unit tests in `internal/resources/clustershared/helpers_test.go` (package `clustershared`): shared schema attributes per data-model.md §4 "Common cluster schema", attribute-to-endpoint update-routing function (per data-model.md "Update routing ordering"), pure function `RouteUpdate(diff Diff) []Operation` returning the ordered op sequence
- [X] T092 [US2] Implement `internal/resources/clustershared/helpers.go` (pure functions only — no I/O) until T091 passes
- [X] T093 [P] [US2] Write `internal/resources/clustershared/README.md` documenting the shared schema and the update-routing precedence

### Managed resource: `nc2_aws_cluster`

- [ ] T094 [P] [US2] Author failing schema-shape contract test in `internal/resources/aws_cluster/schema_test.go` asserting the schema matches `contracts/resources/nc2_aws_cluster.json`, including the AWS-only `access_policy` attribute
- [ ] T095 [P] [US2] Author failing CRUD unit tests in `internal/resources/aws_cluster/resource_test.go` using httptest: Create via `POST /clusters/aws` returning 202 + task id → poll until `done` → Read; Read via `GET /clusters/{id}` with 404 → state removal (FR-009); Delete via `POST /clusters/{id}/terminate` + poll
- [ ] T096 [P] [US2] Author failing in-place-update unit tests in `internal/resources/aws_cluster/update_test.go`: capacity-diff → `update-capacity`; ssh-key-diff → `update-ssh-key`; license/aos_version/software_tier diff → `update-license`; resource_tags diff → `update-resource-tags`; access_policy diff → `update-access-policy`; generic mutable field → `PATCH /clusters/{id}`; combined diff applies operations in the ordering from data-model.md §4
- [ ] T097 [P] [US2] Author failing RequiresReplace unit test in `internal/resources/aws_cluster/replace_test.go`: changes to `name`, `region`, `network.mode`, `network.vpc_cidr`, `network.aws.subnets`, `redundancy.factor` produce destroy+recreate plans (US2 scenario 9)
- [ ] T098 [P] [US2] Author failing drift-free + clean-destroy unit tests in `internal/resources/aws_cluster/lifecycle_test.go` (FR-018, FR-019)
- [ ] T099 [P] [US2] Author failing audit-emission unit test in `internal/resources/aws_cluster/audit_test.go` (FR-020d) covering create-with-poll (multiple records), update-with-routing (one per called endpoint), and delete-with-poll
- [ ] T100 [P] [US2] Author failing ImportState unit test in `internal/resources/aws_cluster/import_test.go` asserting clean plan after import (FR-013, SC-009)
- [ ] T101 [US2] Implement `internal/resources/aws_cluster/schema.go` (extends `clustershared` with AWS-only fields) until T094 passes
- [ ] T102 [US2] Implement `internal/resources/aws_cluster/sensitive.go` per data-model.md §4 (access-policy ip_addresses included)
- [ ] T103 [US2] Implement `internal/resources/aws_cluster/resource.go` (Create/Read/Update with diff-routing/Delete/ImportState; integrates `client.PollTask`) until T095..T100 pass
- [ ] T104 [US2] Implement `internal/resources/aws_cluster/openapi_mapping.go` covering all 8 AWS-applicable cluster operations from data-model.md §4 (excluding hibernate/resume — those land in US4)
- [ ] T105 [P] [US2] Write `internal/resources/aws_cluster/README.md` covering the AWS-only access policy (FR-010a) prominently
- [ ] T106 [US2] Acceptance test in `tests/acceptance/nc2_aws_cluster_test.go` (FR-023): full lifecycle on a sandbox AWS cloud account, including the in-place-update steps for capacity, ssh-key, license, tags, access-policy

### Managed resource: `nc2_azure_cluster`

- [ ] T107 [P] [US2] Author failing schema-shape unit tests in `internal/resources/azure_cluster/schema_test.go` asserting the schema mirrors AWS minus AWS-only fields, plus Azure-only `network.azure.*` per data-model.md §4
- [ ] T108 [P] [US2] Author failing FR-010a rejection test in `internal/resources/azure_cluster/policy_reject_test.go`: setting `access_policy` produces a plan-time validation error (US2 scenario 8)
- [ ] T109 [P] [US2] Author failing CRUD + update-routing + lifecycle + audit unit tests in `internal/resources/azure_cluster/{resource_test.go,update_test.go,lifecycle_test.go,audit_test.go}` analogous to T095..T099 minus access-policy
- [ ] T110 [US2] Implement `internal/resources/azure_cluster/{schema.go,sensitive.go,resource.go,openapi_mapping.go,README.md}` until T107..T109 pass
- [ ] T111 [US2] Acceptance test in `tests/acceptance/nc2_azure_cluster_test.go` (FR-023)

### Managed resource: `nc2_gcp_cluster`

- [ ] T112 [P] [US2] Author failing schema-shape unit tests in `internal/resources/gcp_cluster/schema_test.go` asserting the schema mirrors AWS minus AWS-only fields, plus GCP-only `network.gcp.*` per data-model.md §4
- [ ] T113 [P] [US2] Author failing FR-010a rejection test in `internal/resources/gcp_cluster/policy_reject_test.go`
- [ ] T114 [P] [US2] Author failing CRUD + update-routing + lifecycle + audit unit tests in `internal/resources/gcp_cluster/*_test.go`
- [ ] T115 [US2] Implement `internal/resources/gcp_cluster/{schema.go,sensitive.go,resource.go,openapi_mapping.go,README.md}` until T112..T114 pass
- [ ] T116 [US2] Acceptance test in `tests/acceptance/nc2_gcp_cluster_test.go` (FR-023)

### Data sources for US2

- [ ] T117 [P] [US2] Author failing unit tests for `data.nc2_clusters` (list) in `internal/datasources/clusters/datasource_test.go`: deterministic ordering (FR-015), audit record emitted
- [ ] T118 [P] [US2] Author failing unit tests for `data.nc2_cluster` (single by id) in `internal/datasources/cluster/datasource_test.go`
- [ ] T119 [US2] Implement `internal/datasources/clusters/{datasource.go,schema.go,openapi_mapping.go,README.md}` until T117 passes
- [ ] T120 [US2] Implement `internal/datasources/cluster/{datasource.go,schema.go,openapi_mapping.go,README.md}` until T118 passes
- [ ] T121 [P] [US2] Acceptance test for both cluster data sources in `tests/acceptance/nc2_clusters_datasources_test.go` (FR-024)

### Wire-up + examples + docs for US2

- [ ] T122 [US2] Wire all 3 cluster resources and 2 cluster data sources into `internal/provider/provider.go`
- [ ] T123 [P] [US2] Add runnable examples under `examples/resources/nc2_aws_cluster/`, `examples/resources/nc2_azure_cluster/`, `examples/resources/nc2_gcp_cluster/` (FR-029, SC-006) — each example references US1 resources for org / cloud_account / region
- [ ] T124 [P] [US2] Add runnable examples for `nc2_clusters` and `nc2_cluster` data sources under `examples/data-sources/nc2_clusters/main.tf` and `examples/data-sources/nc2_cluster/main.tf`
- [ ] T125 [P] [US2] Generate docs pages under `docs/resources/nc2_<cloud>_cluster.md` and `docs/data-sources/nc2_cluster{,s}.md` via `tfplugindocs` (FR-030)
- [ ] T126 [US2] Update `internal/oapi` and re-run `make openapi-coverage` — cluster CRUD + in-place-update operations move from uncovered to covered

**Checkpoint**: User Stories 1 + 2 work independently. SC-004 (full org → cloud account → cluster journey on any of 3 clouds) is demonstrable end-to-end. STOP HERE for second internal demo / beta tag.

---

## Phase 5: User Story 3 — Inventory and discovery via data sources (Priority: P2)

**Goal**: Deliver the remaining 10 read-only data sources: availability zones, SSH keys, Prism Centrals, VNets, VPCs, remote storage profiles, cluster cloud resources, notifications, tasks (list), task (single).

**Independent Test**: Per spec US3 — each data source returns results consistent with NC2 API for the same query parameters, stable ordering across reads.

**Depends on**: Phase 2 only (read-only, no managed-resource dependency). Can run in parallel with Phase 4.

- [ ] T127 [P] [US3] Author failing unit tests for `data.nc2_availability_zones` in `internal/datasources/availability_zones/datasource_test.go`: `GET /cloud-accounts/{cloud_account_id}/regions/{region_id}/availability-zones`, deterministic ordering, audit record
- [ ] T128 [P] [US3] Author failing unit tests for `data.nc2_ssh_keys` in `internal/datasources/ssh_keys/datasource_test.go`
- [ ] T129 [P] [US3] Author failing unit tests for `data.nc2_prism_centrals` in `internal/datasources/prism_centrals/datasource_test.go`
- [ ] T130 [P] [US3] Author failing unit tests for `data.nc2_vnets` in `internal/datasources/vnets/datasource_test.go`
- [ ] T131 [P] [US3] Author failing unit tests for `data.nc2_vpcs` in `internal/datasources/vpcs/datasource_test.go`
- [ ] T132 [P] [US3] Author failing unit tests for `data.nc2_remote_storage_profiles` in `internal/datasources/remote_storage_profiles/datasource_test.go`
- [ ] T133 [P] [US3] Author failing unit tests for `data.nc2_cluster_cloud_resources` in `internal/datasources/cluster_cloud_resources/datasource_test.go`
- [ ] T134 [P] [US3] Author failing unit tests for `data.nc2_notifications` in `internal/datasources/notifications/datasource_test.go`
- [ ] T135 [P] [US3] Author failing unit tests for `data.nc2_tasks` (list) in `internal/datasources/tasks/datasource_test.go`
- [ ] T136 [P] [US3] Author failing unit tests for `data.nc2_task` (single) in `internal/datasources/task/datasource_test.go`
- [ ] T137 [US3] Implement `internal/datasources/availability_zones/{datasource.go,schema.go,openapi_mapping.go,README.md}` until T127 passes
- [ ] T138 [US3] Implement `internal/datasources/ssh_keys/...` until T128 passes
- [ ] T139 [US3] Implement `internal/datasources/prism_centrals/...` until T129 passes
- [ ] T140 [US3] Implement `internal/datasources/vnets/...` until T130 passes
- [ ] T141 [US3] Implement `internal/datasources/vpcs/...` until T131 passes
- [ ] T142 [US3] Implement `internal/datasources/remote_storage_profiles/...` until T132 passes
- [ ] T143 [US3] Implement `internal/datasources/cluster_cloud_resources/...` until T133 passes
- [ ] T144 [US3] Implement `internal/datasources/notifications/...` until T134 passes
- [ ] T145 [US3] Implement `internal/datasources/tasks/...` until T135 passes
- [ ] T146 [US3] Implement `internal/datasources/task/...` until T136 passes
- [ ] T147 [P] [US3] Acceptance tests in `tests/acceptance/nc2_us3_datasources_test.go` (FR-024): one sub-test per data source, gated by TF_ACC
- [ ] T148 [US3] Wire all 10 US3 data sources into `internal/provider/provider.go`
- [ ] T149 [P] [US3] Add runnable examples under `examples/data-sources/<each>/` (FR-029, SC-006)
- [ ] T150 [P] [US3] Generate docs pages under `docs/data-sources/<each>.md` via `tfplugindocs` (FR-030)
- [ ] T151 [US3] Re-run `make openapi-coverage` — all data-source operations now covered

**Checkpoint**: All inventory data sources work. Users can drive cluster definitions from live NC2 inventory.

---

## Phase 6: User Story 4 — Hibernate / resume via `desired_state` (Priority: P3)

**Goal**: Extend the 3 cluster resources from US2 with a `desired_state` attribute (`running` | `hibernated`) routed to `POST /clusters/{id}/hibernate` and `POST /clusters/{id}/resume`.

**Independent Test**: Per spec US4 — flip `desired_state` to `hibernated`, apply, plan-clean; flip back to `running`, apply, plan-clean.

**Depends on**: Phase 4 (US2 cluster resources). Touches files in `internal/resources/aws_cluster/`, `azure_cluster/`, `gcp_cluster/` but does not break their existing tests.

- [X] T152 [P] [US4] Author failing unit tests in `internal/resources/clustershared/desired_state_test.go`: pure function `func MapDesiredStateTransition(prev, next string) (Op, error)` returning `OpHibernate`, `OpResume`, or `OpNoop`; rejects invalid transitions; rejects "running" → "running" as no-op
- [X] T153 [US4] Implement `internal/resources/clustershared/desired_state.go` until T152 passes
- [ ] T154 [P] [US4] Author failing AWS hibernate/resume unit tests in `internal/resources/aws_cluster/hibernate_test.go`: `desired_state` diff drives the matching endpoint after all other update operations (per the ordering in data-model.md §4); plan-after-apply is clean; rejecting "hibernated" + "capacity change" in the same diff at plan time
- [ ] T155 [P] [US4] Author failing Azure hibernate/resume unit tests in `internal/resources/azure_cluster/hibernate_test.go` mirroring T154
- [ ] T156 [P] [US4] Author failing GCP hibernate/resume unit tests in `internal/resources/gcp_cluster/hibernate_test.go` mirroring T154
- [ ] T157 [US4] Extend `internal/resources/aws_cluster/resource.go` Update method to dispatch hibernate/resume after the existing ordering until T154 passes
- [ ] T158 [US4] Extend `internal/resources/azure_cluster/resource.go` analogously until T155 passes
- [ ] T159 [US4] Extend `internal/resources/gcp_cluster/resource.go` analogously until T156 passes
- [ ] T160 [US4] Update each cluster's `openapi_mapping.go` to include hibernate + resume operationIds
- [ ] T161 [US4] Acceptance test in `tests/acceptance/nc2_cluster_hibernate_resume_test.go` (FR-023): for each cloud, apply running → flip to hibernated → plan-clean → flip back → plan-clean
- [ ] T162 [P] [US4] Add runnable example `examples/resources/nc2_aws_cluster/hibernate.tf` (and equivalents for Azure / GCP) showing the `desired_state` toggle (FR-029)
- [ ] T163 [P] [US4] Update generated docs for the 3 cluster resources to call out the `desired_state` attribute and the hibernate/resume transition semantics (FR-030)
- [ ] T164 [US4] Re-run `make openapi-coverage` — `hibernate` and `resume` operations now covered

**Checkpoint**: Cost-optimization workflow lands. SC-002 / SC-003 continue to pass on all cluster types in both running and hibernated states.

---

## Phase 7: User Story 5 — Non-CRUD operational endpoints as Terraform actions (Priority: P3)

**Goal**: Deliver 8 actions: `nc2_cluster_condemn_host`, `nc2_cluster_open_support_tunnel`, `nc2_cluster_extend_support_tunnel`, `nc2_cluster_close_support_tunnel`, `nc2_cluster_scale_flow_gateway`, `nc2_cluster_upgrade_flow_gateway`, `nc2_cluster_start_recovery`, `nc2_notification_acknowledge`.

**Independent Test**: Per spec US5 — for each action, invoke against the prerequisite (cluster or notification), verify the right NC2 endpoint is called and the task / response is surfaced as the action's structured result.

**Depends on**: Phase 4 (US2 cluster resources) for cluster-scoped actions; Phase 5 (US3 notifications data source) for `nc2_notification_acknowledge`.

### Shared action scaffolding

- [ ] T165 [US5] Author failing unit tests in `internal/actions/shared/result_test.go`: shared `Result` object schema per data-model.md §11 (task_id, status, http_status, error_code, error_message), pure function `func ResultFromTask(t TaskResult) Result`
- [ ] T166 [US5] Implement `internal/actions/shared/result.go` until T165 passes

### Per-action test+impl pairs

- [ ] T167 [P] [US5] Author failing unit tests in `internal/actions/cluster_condemn_host/action_test.go` (contract test against `contracts/actions/nc2_cluster_condemn_host.json`): Invoke calls `POST /clusters/{id}/condemn-host` with `data.host_id`, polls task, exposes `result.*` (FR-017); audit record emitted per call (FR-020d)
- [ ] T168 [P] [US5] Author failing unit tests in `internal/actions/cluster_open_support_tunnel/action_test.go`
- [ ] T169 [P] [US5] Author failing unit tests in `internal/actions/cluster_extend_support_tunnel/action_test.go` (with `duration_hours` input)
- [ ] T170 [P] [US5] Author failing unit tests in `internal/actions/cluster_close_support_tunnel/action_test.go`
- [ ] T171 [P] [US5] Author failing unit tests in `internal/actions/cluster_scale_flow_gateway/action_test.go` (with `target_node_count` input)
- [ ] T172 [P] [US5] Author failing unit tests in `internal/actions/cluster_upgrade_flow_gateway/action_test.go`
- [ ] T173 [P] [US5] Author failing unit tests in `internal/actions/cluster_start_recovery/action_test.go`
- [ ] T174 [P] [US5] Author failing unit tests in `internal/actions/notification_acknowledge/action_test.go`: PATCH `/notifications/{id}` with `{"data":{"acknowledged":true}}` (US5 scenario 5)
- [ ] T175 [US5] Implement `internal/actions/cluster_condemn_host/{action.go,schema.go,openapi_mapping.go,README.md}` until T167 passes
- [ ] T176 [US5] Implement `internal/actions/cluster_open_support_tunnel/...` until T168 passes
- [ ] T177 [US5] Implement `internal/actions/cluster_extend_support_tunnel/...` until T169 passes
- [ ] T178 [US5] Implement `internal/actions/cluster_close_support_tunnel/...` until T170 passes
- [ ] T179 [US5] Implement `internal/actions/cluster_scale_flow_gateway/...` until T171 passes
- [ ] T180 [US5] Implement `internal/actions/cluster_upgrade_flow_gateway/...` until T172 passes
- [ ] T181 [US5] Implement `internal/actions/cluster_start_recovery/...` until T173 passes
- [ ] T182 [US5] Implement `internal/actions/notification_acknowledge/...` until T174 passes

### Acceptance tests + wire-up

- [ ] T183 [P] [US5] Acceptance test in `tests/acceptance/nc2_us5_cluster_actions_test.go` (FR-025): one sub-test per cluster-scoped action against a sandbox cluster
- [ ] T184 [P] [US5] Acceptance test in `tests/acceptance/nc2_notification_acknowledge_test.go` (FR-025)
- [ ] T185 [US5] Wire all 8 actions into `internal/provider/provider.go` `Actions()` slice
- [ ] T186 [P] [US5] Add runnable examples under `examples/actions/<each>/main.tf` (FR-029, SC-006)
- [ ] T187 [P] [US5] Generate docs pages under `docs/actions/<each>.md` via `tfplugindocs` (FR-030)
- [ ] T188 [US5] Re-run `make openapi-coverage` — all 8 action operations now covered. Coverage report MUST now show 49 / 49 covered (SC-001).

**Checkpoint**: 100% OpenAPI coverage. SC-001 met. All user stories independently functional.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, release pipeline, supply-chain attestation, final quality bars across all user stories. No new behavior — every test added here is a regression-guard.

- [ ] T189 [P] Author failing unit tests in `internal/redact/integration_test.go` aggregating across every per-package `Sensitive()` registry to assert no plaintext escapes (FR-002b at scale)
- [ ] T190 [P] Author failing integration test in `tests/integration/coverage_test.go` asserting `tools/coverage-check` exits 0 against the committed OpenAPI and that the produced `coverage-report.json` has `uncovered: []`
- [ ] T191 [P] Author failing test in `tests/integration/no_skip_verify_test.go` (FR-003c at the binary level): `go build` then run a binary-symbol grep for `InsecureSkipVerify: true` patterns; CI fails if any are found
- [ ] T192 [P] Author failing test in `tests/integration/no_external_secret_stores_test.go` (FR-001b at the binary level): `go list -m all` does not contain any forbidden module path (vault, secretsmanager, keyvault, secret-manager, 1password)
- [ ] T193 [P] Author failing docs-coverage test in `tests/integration/docs_coverage_test.go`: every resource / data source / action under `internal/{resources,datasources,actions}/` has a corresponding `docs/{resources,data-sources,actions}/<type>.md` and at least one runnable example under `examples/...` (FR-029, SC-006)
- [ ] T194 [P] Author failing doc-comment test in `tests/integration/godoc_coverage_test.go` invoking `revive --rules exported` and asserting zero violations (SC-007)
- [ ] T195 Implement `tests/integration/*_test.go` infrastructure (helpers, fixtures) until T189..T194 pass
- [ ] T196 [P] Write `docs/index.md` — provider overview, link tree, security posture summary
- [ ] T197 [P] Write `docs/guides/getting-started.md` — end-user (not developer) onramp
- [ ] T198 [P] Write `docs/guides/authentication.md` — credential resolution (FR-001a), credentials file format, profile selection, conflict warning, no external-secret-store imports (FR-001b)
- [ ] T199 [P] Write `docs/guides/security-hardening.md` — TLS posture (FR-003b/c), `ca_bundle`, sensitive-field redaction (FR-002), audit log routing (FR-020c), release signing (FR-032), supply-chain attestation (FR-032b)
- [ ] T200 [P] Write `docs/guides/async-and-tasks.md` — async task model (FR-003, FR-008), timeout configuration, troubleshooting via task IDs and audit records
- [ ] T201 [P] Add `CHANGELOG.md` in Keep-a-Changelog format with the first entry `## [Unreleased]` (FR-031)
- [ ] T202 Author failing release-pipeline test in `.github/workflows/release.test.yml` (using `act` or a dry-run dispatch path): asserts a tag dispatch triggers GoReleaser, cosign keyless signatures, SLSA L3 attestation, GPG signing of SHA256SUMS, govulncheck, osv-scanner (FR-032, FR-032a, FR-032b)
- [ ] T203 Implement `.goreleaser.yml` with cross-compile matrix (linux/darwin/windows/freebsd × amd64/arm64/386 per plan.md) until the release dry-run from T202 succeeds
- [ ] T204 Implement `.github/workflows/release.yml` orchestrating: GoReleaser → cosign-sign-blob (keyless) → slsa-github-generator → GPG-sign SHA256SUMS → govulncheck → osv-scanner; HIGH/CRITICAL CVEs HARD-FAIL the workflow (FR-032a); on success publish artifacts: provider binaries, `SHA256SUMS`, `SHA256SUMS.sig`, per-artifact `.sig`/`.pem`, `provenance.intoto.jsonl`, `vulnerability-report.json` (FR-032b)
- [ ] T205 Add `terraform-registry-manifest.json` declaring the provider's `protocol_versions = ["6.0"]` for `terraform-plugin-framework` 1.x (FR-032)
- [ ] T206 Final OpenAPI coverage gate: run `make openapi-coverage` and confirm `coverage-report.json` shows `total_operations: 49, covered: 49, uncovered: []` (SC-001)
- [ ] T207 Final unit-test coverage gate: run `make coverage` and confirm ≥ 80% line coverage on `internal/...` packages (SC-005)
- [ ] T208 Final `make doc-lint` run confirms 100% of exported Go symbols have doc comments (SC-007)
- [ ] T209 Final `make sensitive-lint-strict` run confirms every FR-002-pattern-matching attribute in `openapi/openapi.json` is classified in a per-resource registry (FR-002a)
- [ ] T210 Final `make vuln` run confirms zero HIGH or CRITICAL CVEs in the dependency graph (FR-032a)
- [ ] T211 Run `quickstart.md` §6 end-to-end against a sandbox NC2 tenant: provision org + AWS cloud account → verify drift-free plan and clean destroy; SC-004 step (a) measured in minutes
- [ ] T212 Run full SC-004 journey: provision org + cloud account + 1 cluster on each of AWS, Azure, GCP; record elapsed time; assert ≤ 120 min total, ≤ 5 min provider overhead

**Checkpoint**: All success criteria SC-001 through SC-010 measurable and green. Ready to tag v0.1.0 and publish to the Terraform Registry.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)** — no dependencies; can start immediately.
- **Phase 2 (Foundational)** — depends on Phase 1; BLOCKS every user-story phase.
- **Phase 3 (US1 / P1)** — depends on Phase 2; this is the MVP.
- **Phase 4 (US2 / P2)** — depends on Phase 3 (cluster needs org + cloud account + region).
- **Phase 5 (US3 / P2)** — depends on Phase 2 only; CAN RUN IN PARALLEL with Phase 4.
- **Phase 6 (US4 / P3)** — depends on Phase 4 (extends cluster resources).
- **Phase 7 (US5 / P3)** — depends on Phase 4 (cluster-scoped actions) and Phase 5 (`nc2_notification_acknowledge` references notifications); within Phase 7, all 8 actions can run in parallel.
- **Phase 8 (Polish)** — depends on whichever user stories are in scope for the release.

### User Story Dependencies

- US1 → US2 (cluster needs org + cloud account + region from US1)
- US2 → US4 (hibernate/resume extends cluster resources)
- US2 + US5 (cluster-scoped actions depend on cluster managed resources)
- US3 + US5 (`nc2_notification_acknowledge` references `nc2_notifications` data source from US3)
- US3 is parallelizable with US2 (read-only, no shared state)

### Within Each User Story

- TDD: failing test task FIRST, implementation task SECOND (Constitution Principle II, NON-NEGOTIABLE).
- Schema → sensitive registry → resource/data-source/action body → OpenAPI mapping → README → wire-up → acceptance test → examples → docs.

### Parallel Opportunities

- All [P] tasks in Phase 1 can run in parallel.
- All [P] tasks in Phase 2 within a single library (e.g., redact, auth) can run in parallel; libraries themselves are independent and can be staffed in parallel.
- Within Phase 3, the 3 managed resources (`nc2_organization`, `nc2_cloud_account`, `nc2_cloud_account_region`) can be staffed in parallel after T086 wire-up coordination.
- Within Phase 4, the 3 cluster resources (`nc2_aws_cluster`, `nc2_azure_cluster`, `nc2_gcp_cluster`) can be staffed in parallel after `clustershared` (T091/T092) lands.
- All 10 data source test-tasks T127–T136 in Phase 5 can run in parallel (different files, no dependencies).
- All 8 action test-tasks T167–T174 in Phase 7 can run in parallel.
- All [P] documentation/example tasks in Phase 8 can run in parallel.

---

## Parallel Example: User Story 1 MVP — kickoff after Phase 2 completes

```bash
# Tests first (Constitution Principle II), all in parallel:
Task: "T046 Author failing schema-shape test for nc2_organization in internal/resources/organization/schema_test.go"
Task: "T047 Author failing CRUD test for nc2_organization in internal/resources/organization/resource_test.go"
Task: "T057 Author failing schema-shape test for nc2_cloud_account in internal/resources/cloud_account/schema_test.go"
Task: "T058 Author failing CRUD test for nc2_cloud_account in internal/resources/cloud_account/resource_test.go"
Task: "T068 Author failing schema-shape test for nc2_cloud_account_region in internal/resources/cloud_account_region/schema_test.go"
Task: "T073 Author failing test for data.nc2_organizations in internal/datasources/organizations/datasource_test.go"

# Then implementations in parallel (after their respective tests are red and committed):
Task: "T051+T052+T053+T054 Implement nc2_organization (schema, sensitive, resource, openapi_mapping)"
Task: "T062+T063+T064+T065 Implement nc2_cloud_account (schema, sensitive, resource, openapi_mapping)"
Task: "T071 Implement nc2_cloud_account_region (schema + sensitive + resource + openapi_mapping)"
Task: "T079..T084 Implement US1 data sources"
```

---

## Implementation Strategy

### MVP First (User Story 1 only)

1. Complete Phase 1: Setup.
2. Complete Phase 2: Foundational (CRITICAL — blocks all stories).
3. Complete Phase 3: User Story 1 (`nc2_organization` + `nc2_cloud_account` + `nc2_cloud_account_region` + 6 data sources).
4. **STOP and VALIDATE**: run the US1 acceptance suite + the US1 portion of `quickstart.md` §6 against a sandbox.
5. Tag `v0.0.1-alpha` and demo. The provider is already useful for IaC platform owners onboarding NC2 tenants.

### Incremental Delivery

1. Setup + Foundational → toolchain green.
2. US1 (MVP) → tag `v0.0.1-alpha`, internal demo.
3. US2 + US3 (parallel) → tag `v0.0.5-beta`, external demo on all three clouds.
4. US4 → tag `v0.0.8-rc`, cost-optimization workflow live.
5. US5 → tag `v0.0.9-rc`, full OpenAPI coverage.
6. Phase 8 polish → tag `v0.1.0`, publish to Terraform Registry with full FR-032 signing + provenance.

### Parallel Team Strategy

With 3 developers:

1. Together: Phase 1 + Phase 2 (estimated 2–3 weeks).
2. Once Phase 2 is green:
   - Developer A: Phase 3 (US1 — MVP critical path).
   - Developer B: queue waiting for Phase 3, in the meantime Phase 5 (US3 — independent).
   - Developer C: Phase 8 cross-cutting items that don't require resources to exist (release workflow, docs scaffolding, security tests).
3. After Phase 3 lands:
   - Developer A: Phase 4 (US2 — clusters).
   - Developer B: continues Phase 5.
   - Developer C: Phase 8 docs + release.
4. Phase 6 + Phase 7 fold in after Phase 4 + Phase 5 complete; both have a small footprint and can be split among A/B/C.

---

## Notes

- [P] tasks operate on different files with no incomplete-task dependencies.
- [Story] labels (US1..US5) map every story-phase task back to its user story for traceability; Setup, Foundational, and Polish phases have no story label.
- TDD is NON-NEGOTIABLE (Constitution Principle II): write the failing test, observe its red, then implement. Commits should ideally have a test+impl pair.
- Every behavior change task ships its test in the same diff. The "test-required" CI gate (FR-027) blocks merges otherwise.
- The OpenAPI coverage gate (FR-021, T206) is the single most important durable invariant — if the API changes upstream and this provider is rebuilt, the gate will catch new operations before they ship silently uncovered.
- The two no-DELETE caveats (cloud account FR-006, cloud account region FR-007) MUST surface as plan-time warnings; this is captured in T058–T059 and T069.
- Avoid: vague tasks ("implement clusters"), same-file conflicts in parallel work (each [P] task targets a distinct file), cross-story dependencies that break independence (US3 is intentionally parallelizable with US2).
