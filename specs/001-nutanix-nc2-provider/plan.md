# Implementation Plan: Terraform Provider for Nutanix Cloud Clusters (NC2)

**Branch**: `001-nutanix-nc2-provider` | **Date**: 2026-05-13 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/001-nutanix-nc2-provider/spec.md`

**Authoritative API**: `openapi/openapi.json` (committed) — 49 operations across 40 paths, 13 tag groups.

## Summary

Ship a Terraform provider, `terraform-provider-nc2`, that exposes 100% of the Nutanix Cloud Clusters v2 API surface as a clean Terraform-native experience: 6 managed resources, 18 data sources, and 8 actions, with strict drift-free / clean-destroy quality bars on every managed resource and an OpenAPI-driven coverage gate to prevent silent fall-behind.

The technical approach is conservative and security-forward:

- Built on `terraform-plugin-framework` (1.x, current GA), using its first-class `actions` concept for non-CRUD operational endpoints (FR-016) and its `tflog` facility for the per-call structured audit log (FR-020a..d).
- Written in Go (1.24+) with a deliberately small dependency graph — standard library `net/http` and `encoding/json` over any third-party HTTP / serialization library, to keep the supply-chain attack surface narrow and to satisfy FR-032 / FR-032a (HIGH/CRITICAL CVE release gate).
- Internal code organized as small standalone libraries (`auth`, `client`, `audit`, `redact`, schema helpers) per Constitution Principle I (Library-First); each has its own README, tests, and clear single purpose.
- TDD enforced via CI gates and a pre-commit hook (Constitution Principle II): no production change merges without a failing-then-passing test in the same diff.
- Pure functional patterns by default (Principle III): credential resolution, JWT minting (given inputs), schema translation, audit-record formatting, and sensitive-field redaction are all pure; I/O is isolated to a single HTTP client struct and the framework's logging edge.
- Release shipping via GoReleaser, signed with GPG + Sigstore cosign keyless signatures + SLSA Build L3 provenance (FR-032), gated by `govulncheck` and `osv-scanner` (FR-032a).

The MVP slice (US1) delivers organizations + cloud accounts; the next slice (US2) delivers full cluster CRUD on all three clouds; US3 (inventory data sources) is parallelizable with US2; US4 (hibernate/resume) and US5 (actions) follow.

## Technical Context

**Language/Version**: Go 1.24+ (latest stable supported by `terraform-plugin-framework`). Pinned in `go.mod`; CI matrix covers Go 1.24 and 1.23.

**Primary Dependencies**:

- `github.com/hashicorp/terraform-plugin-framework` — the latest plugin framework (1.x), the user's explicit "latest terraform provider framework" choice. Provides Resource, DataSource, Action, EphemeralResource, schema, plan modifiers, validators, and `tflog`.
- `github.com/hashicorp/terraform-plugin-framework-validators` — schema validators (string-in-list, length, regex).
- `github.com/hashicorp/terraform-plugin-log` — already pulled by the framework, re-used directly for the FR-020a..d audit emitter.
- `github.com/hashicorp/terraform-plugin-testing` — the official acceptance test harness for the framework (replaces the legacy SDKv2 testing module).
- `github.com/golang-jwt/jwt/v5` — JWT minting for FR-001 (HS512, custom `kid` header, MyNutanix audience).
- `github.com/getkin/kin-openapi` (tools-only, behind a `// +build tools` constraint) — used only by the `tools/coverage-check` and `tools/sensitive-lint` CI binaries to parse `openapi/openapi.json`. **Not** linked into the provider binary, keeping the runtime dependency graph tiny.
- Standard library only for `net/http`, `encoding/json`, `crypto/tls`, `time`, `context`. No third-party HTTP, JSON, or retry library.
- Test-only: `github.com/stretchr/testify` (assertion helpers); `net/http/httptest` (fake NC2 server).
- Tools (CI-only, not linked into binary): `golang.org/x/vuln/cmd/govulncheck`, `github.com/google/osv-scanner/cmd/osv-scanner`, `github.com/sigstore/cosign/v2/cmd/cosign`, `github.com/slsa-framework/slsa-github-generator`, `github.com/goreleaser/goreleaser`.

**Storage**: N/A at the provider level. Terraform state persistence is the user's responsibility (managed by Terraform itself); the provider never touches a state file directly. The optional `credentials_file` (FR-001a, default `~/.nc2/credentials`) is read-only and lives on the operator's workstation, not in the provider's domain.

**Testing**:

- **Unit tests**: Go stdlib `testing` + `testify/require`, with `net/http/httptest` standing in for the NC2 API. Every package has unit tests; coverage gate ≥ 80% (SC-005). Unit tests run on every PR and never hit a real NC2 endpoint.
- **Acceptance tests**: `github.com/hashicorp/terraform-plugin-testing` with `TF_ACC=1`. Run on demand via `make testacc` and on a scheduled CI lane. Gated by `NC2_API_KEY` + `NC2_KEY_ID` + `NC2_ISSUER` env vars; absent credentials means tests are skipped, not failing.
- **Coverage report**: a `tools/coverage-check` binary parses `openapi/openapi.json` and the generated registry of resources/data sources/actions; fails CI if any operation is unmapped (FR-021, FR-021a).
- **Sensitive-field lint**: a `tools/sensitive-lint` binary scans the OpenAPI for new fields matching FR-002 patterns and cross-checks against per-resource sensitive registries; fails CI in `strict` mode for unclassified hits (FR-002a).

**Target Platform**: Cross-compiled to all platforms supported by the Terraform Registry: `linux_amd64`, `linux_arm64`, `linux_386`, `darwin_amd64`, `darwin_arm64`, `windows_amd64`, `windows_386`, `freebsd_amd64`, `freebsd_arm`. Built by GoReleaser in a single CI run.

**Project Type**: Terraform provider (plugin / library). Distributed through the Terraform Registry; consumed by users via `terraform init` resolving the provider source `nutanix/nc2` (namespace TBD at first release).

**Performance Goals**:

- Provider startup overhead ≤ 100 ms wall clock (cold).
- JWT minting ≤ 5 ms (HS512 over ~200 bytes).
- Steady-state per-call overhead added by the provider (auth + audit + JSON marshalling) ≤ 10 ms p95 outside the NC2 server's response latency.
- Cumulative provider-induced overhead in the SC-004 end-to-end journey ≤ 5 minutes (the user-facing budget; the bulk of the 120 min is NC2 provisioning latency).

**Constraints**:

- No `InsecureSkipVerify` flag in any build (FR-003b) — verified by FR-003c unit test.
- No native external-secret-store imports (FR-001b) — verified by an import-graph check in CI.
- 5-minute JWT lifetime with transparent refresh (FR-001).
- 100% OpenAPI operation coverage (FR-021, SC-001).
- ≥ 80% unit test line coverage (SC-005).
- 100% exported Go symbols have doc comments (SC-007) — enforced by `revive` rule `exported` in CI.
- HIGH/CRITICAL CVE release gate (FR-032a).

**Scale/Scope**:

- **Managed resources** (6): `nc2_organization`, `nc2_cloud_account`, `nc2_cloud_account_region`, `nc2_aws_cluster`, `nc2_azure_cluster`, `nc2_gcp_cluster`.
- **Data sources** (18): `nc2_organizations`, `nc2_organization`, `nc2_organization_audit_trail`, `nc2_cloud_accounts`, `nc2_cloud_account`, `nc2_cloud_account_regions`, `nc2_availability_zones`, `nc2_ssh_keys`, `nc2_prism_centrals`, `nc2_vnets`, `nc2_vpcs`, `nc2_remote_storage_profiles`, `nc2_clusters`, `nc2_cluster`, `nc2_cluster_cloud_resources`, `nc2_notifications`, `nc2_tasks`, `nc2_task`.
- **Actions** (8): `nc2_cluster_condemn_host`, `nc2_cluster_open_support_tunnel`, `nc2_cluster_extend_support_tunnel`, `nc2_cluster_close_support_tunnel`, `nc2_cluster_scale_flow_gateway`, `nc2_cluster_upgrade_flow_gateway`, `nc2_cluster_start_recovery`, `nc2_notification_acknowledge`.
- **OpenAPI operations to cover**: 49 across 40 paths (FR-021).
- Estimated codebase size at first release: ~25–35 kLOC Go (provider + tests + tools), ~3–5 kLOC HCL examples, ~50–80 documentation pages (mostly schema-generated).

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

The project constitution is at `.specify/memory/constitution.md` (v1.0.0). It defines three non-negotiable principles and a set of cross-cutting constraints. Each is evaluated below against the planned design.

### Principle I — Library-First (NON-NEGOTIABLE)

**Requirement**: Every feature starts as a self-contained, independently testable library with a single clear purpose and a documented public surface.

**Plan compliance**:

- `internal/auth`, `internal/client`, `internal/audit`, `internal/redact`, and `internal/oapi` (OpenAPI parsing for the CI tools) are standalone libraries. Each has its own `README.md`, unit tests, exported public API, and zero direct dependencies on the others except where explicitly designed (e.g., `client` depends on `auth` for JWT and `audit` for logging).
- Per-resource, per-data-source, and per-action packages under `internal/resources/*`, `internal/datasources/*`, and `internal/actions/*` each have a single purpose (one Terraform resource type per package) and a README.
- No "organizational" / "junk drawer" libraries. The closest temptation is `internal/util`, which is **not** introduced — utility code is either pushed into the library it serves or rejected as not needed.

**Verdict**: PASS.

### Principle II — Test-Driven Development (NON-NEGOTIABLE)

**Requirement**: Strict Red → Green → Refactor; no production code without a failing-then-passing test in the same diff; test suite green on every commit; no long-lived skipped tests.

**Plan compliance**:

- CI runs `go test ./...` on every PR; merge is blocked on failure.
- A pre-commit hook (committed in `.githooks/pre-commit`) runs `go test ./...` locally to catch issues before push.
- `golangci-lint` is configured with the `paralleltest` and `tparallel` linters; t.Skip is allowed only with an `//nolint:tskip` comment that includes a tracking issue ID, and a CI lint scans for any naked `t.Skip` without that annotation.
- The "test-required" gate (FR-027): a CI step diffs the PR against `main` and fails the build if any new exported Go symbol, managed resource, data source, or action lacks a corresponding `*_test.go` file or test function.
- Acceptance tests follow the same TDD cycle, executed in the scheduled CI lane.

**Verdict**: PASS.

### Principle III — Functional Programming Patterns

**Requirement**: Pure functions by default; mutable shared state avoided / isolated / explicit; composition over inheritance; immutability default; imperative/OO confined to integration edges.

**Plan compliance**:

- **Pure**: credential resolution (block / env / file → resolved struct), JWT minting (struct → token + expiry), schema translation (NC2 JSON → Terraform attribute values), audit record formatting (call metadata → JSON string), sensitive-field redaction (record + registry → redacted record), error mapping (HTTP response → user-facing diagnostic). Each is implemented as a stand-alone function with no side effects and unit-tested with table-driven tests.
- **Effects at the edge**: the only mutable state in the provider is the `http.Client` connection pool (stdlib) and the JWT cache held by `auth.TokenManager` (a small struct guarded by a single sync.RWMutex). The JWT cache is internal to one library and never escapes; the goroutine that refreshes it is started lazily by `Configure()` and stopped via context cancellation.
- **No inheritance**: the Terraform plugin framework uses interfaces (`resource.Resource`, `datasource.DataSource`, `action.Action`); we implement them with small struct values and composition, never embedding for behavior reuse.
- **Immutability**: data passed between layers (e.g., NC2 API response → resource state) is immutable in the functional sense (returned by value, not mutated in place); Go's lack of `const` for structs is mitigated by careful API design (return values, accept values, no pointer aliasing in pure paths).

**Verdict**: PASS.

### Additional Constraints (from `## Additional Constraints` in the constitution)

- **Explicit library surface**: each `internal/<library>/` exports only the symbols it intends to be its public API; `internal/...` enforces this at the Go-module level. Verified by a `go vet` + `golangci-lint` rule.
- **No hidden I/O in pure modules**: enforced by code review and reinforced by package layout — pure libraries have no `net/http` or `os` imports beyond `os.Getenv` (and even that is wrapped in a small adapter).
- **Co-located documentation**: every internal library has a `README.md`. Every managed resource, data source, and action has an external `docs/` page (auto-generated where possible; see FR-030).
- **Tests live with the library**: every package has `*_test.go` files alongside; there is no aggregate-only test suite.

**Verdict**: PASS.

### Development Workflow

- Feature branches → `main`; CI must be green to merge.
- Code review checklist (constitution) is encoded in `.github/PULL_REQUEST_TEMPLATE.md`.
- Refactor PRs leave existing tests unchanged; otherwise the change is treated as a behavior change subject to TDD.

**Verdict**: PASS.

**Overall Constitution Check (pre-Phase-0)**: **PASS** — no violations, no Complexity Tracking entries required.

## Project Structure

### Documentation (this feature)

```text
specs/001-nutanix-nc2-provider/
├── plan.md                      # This file
├── spec.md                      # Feature specification (already exists)
├── research.md                  # Phase 0 output (this command)
├── data-model.md                # Phase 1 output (this command)
├── quickstart.md                # Phase 1 output (this command)
├── contracts/                   # Phase 1 output (this command)
│   ├── README.md
│   ├── provider.json
│   ├── resources/
│   │   ├── nc2_organization.json
│   │   └── nc2_aws_cluster.json
│   ├── datasources/
│   │   └── nc2_cloud_account_regions.json
│   └── actions/
│       └── nc2_cluster_condemn_host.json
├── checklists/
│   └── requirements.md          # Spec quality checklist (already exists)
└── tasks.md                     # Phase 2 output — created by /speckit.tasks, NOT here
```

### Source Code (repository root, target layout once implementation starts)

```text
terraform-provider-nc2/                      # repo root, same as this spec-kit project
├── main.go                                  # provider entry; wires terraform-plugin-framework
├── go.mod / go.sum                          # Go 1.24+; minimal dependency set per Technical Context
├── openapi/
│   └── openapi.json                         # authoritative API source (already present)
├── internal/                                # Go-module-enforced internal packages
│   ├── provider/                            # provider definition + config schema
│   │   ├── provider.go
│   │   ├── provider_test.go
│   │   ├── config.go                        # FR-001a credential resolution (block > env > file)
│   │   ├── config_test.go
│   │   ├── tls.go                           # FR-003b/c CA bundle composition (no skip-verify)
│   │   ├── tls_test.go
│   │   └── README.md
│   ├── auth/                                # FR-001 JWT minting + refresh (Library-First)
│   │   ├── jwt.go
│   │   ├── jwt_test.go
│   │   ├── manager.go                       # TokenManager with sync.RWMutex
│   │   ├── manager_test.go
│   │   └── README.md
│   ├── client/                              # NC2 HTTP client + task polling + error mapping
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── tasks.go                         # FR-003 / FR-008 polling loop
│   │   ├── tasks_test.go
│   │   ├── errors.go                        # FR-020 error envelope translation
│   │   ├── errors_test.go
│   │   └── README.md
│   ├── audit/                               # FR-020a..d structured audit emitter
│   │   ├── audit.go
│   │   ├── audit_test.go
│   │   └── README.md
│   ├── redact/                              # FR-002 / FR-002a sensitive-field redaction
│   │   ├── redact.go
│   │   ├── redact_test.go
│   │   ├── patterns.go                      # the FR-002 pattern list
│   │   ├── patterns_test.go
│   │   └── README.md
│   ├── resources/
│   │   ├── organization/                    # nc2_organization (US1)
│   │   │   ├── resource.go
│   │   │   ├── resource_test.go
│   │   │   ├── schema.go
│   │   │   ├── sensitive.go                 # per-resource sensitive registry (FR-002)
│   │   │   └── README.md
│   │   ├── cloud_account/                   # nc2_cloud_account (US1)
│   │   ├── cloud_account_region/            # nc2_cloud_account_region (US1)
│   │   ├── aws_cluster/                     # nc2_aws_cluster (US2)
│   │   ├── azure_cluster/                   # nc2_azure_cluster (US2)
│   │   └── gcp_cluster/                     # nc2_gcp_cluster (US2)
│   ├── datasources/                         # 18 packages, same pattern as resources/
│   │   └── ...
│   └── actions/                             # 8 packages, same pattern
│       └── ...
├── tools/                                   # CI-only Go binaries (behind tools tag)
│   ├── coverage-check/                      # FR-021 / FR-021a OpenAPI coverage gate
│   │   ├── main.go
│   │   └── main_test.go
│   └── sensitive-lint/                      # FR-002a sensitive-field lint
│       ├── main.go
│       └── main_test.go
├── examples/                                # FR-029 runnable examples
│   ├── provider/main.tf
│   ├── resources/nc2_organization/main.tf
│   ├── resources/nc2_cloud_account/main.tf
│   ├── resources/nc2_aws_cluster/main.tf
│   ├── resources/nc2_azure_cluster/main.tf
│   ├── resources/nc2_gcp_cluster/main.tf
│   ├── data-sources/<each data source>/main.tf
│   └── actions/<each action>/main.tf
├── docs/                                    # FR-029 / FR-030 (schema-generated where possible)
│   ├── index.md                             # provider overview
│   ├── guides/
│   │   ├── getting-started.md
│   │   ├── authentication.md
│   │   ├── security-hardening.md
│   │   └── async-and-tasks.md
│   ├── resources/<one md per managed resource>
│   ├── data-sources/<one md per data source>
│   └── actions/<one md per action>
├── tests/
│   ├── acceptance/                          # FR-023..FR-026 E2E with terraform-plugin-testing
│   │   ├── helpers.go
│   │   ├── nc2_organization_test.go
│   │   ├── nc2_cloud_account_test.go
│   │   ├── nc2_aws_cluster_test.go
│   │   └── ...
│   └── fixtures/                            # OpenAPI-derived test fixtures
├── .github/
│   ├── workflows/
│   │   ├── ci.yml                           # unit + lint + coverage + vuln scan on every PR
│   │   ├── acceptance.yml                   # nightly + manual dispatch
│   │   └── release.yml                      # GoReleaser + SLSA L3 + cosign + GPG (FR-032)
│   └── PULL_REQUEST_TEMPLATE.md             # constitution checklist
├── .githooks/
│   └── pre-commit                           # local TDD enforcement
├── .goreleaser.yml                          # GoReleaser config (FR-032)
├── terraform-registry-manifest.json         # FR-032 registry manifest
├── CHANGELOG.md                             # Keep-a-Changelog (FR-031)
├── README.md
└── LICENSE                                  # MPL-2.0 (Terraform-ecosystem convention)
```

**Structure Decision**:

A single Go module rooted at the repository (Option 1 in the template — "single project"). The module is a Terraform provider plugin. Internal libraries (`auth`, `client`, `audit`, `redact`) are first-class citizens with their own READMEs and tests, per the constitution's Library-First principle. CI-only tools live in `tools/` behind a `// +build tools` constraint so they do not pollute the runtime dependency graph. Acceptance tests live in `tests/acceptance/` rather than alongside production code, because they require the `TF_ACC=1` gate, real cloud credentials, and a different CI lane.

## Complexity Tracking

> Filled ONLY if Constitution Check has violations that must be justified.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|--------------------------------------|
| *(none)*  | *(none)*   | *(none)*                             |

The pre-Phase-0 Constitution Check is PASS with no violations. The post-Phase-1 re-check is recorded at the end of the Phase 1 section below.

## Phase 0 — Outline & Research

See `research.md` in this directory. Phase 0 resolved the following research tasks; all results are recorded as Decision / Rationale / Alternatives entries in that file:

1. Confirm the current major version and feature support of `terraform-plugin-framework`, including the maturity of `Action` resources and Action testing.
2. Determine the exact JWT signing recipe required by NC2 (HS512, body = key ID, signature = HMAC-SHA512(api_key, key_id) base64-encoded, header `kid` = key ID, `aud = https://apikeys.nutanix.com`, `exp` ≤ 5 min).
3. Choose the OpenAPI parser for the CI-only tools (`kin-openapi` vs `libopenapi` vs hand-rolled). Confirm it parses `openapi/openapi.json` without errors.
4. Select the cross-compile / signing / provenance tooling: GoReleaser + `slsa-github-generator` + `cosign` + GPG, with Terraform Registry compatibility verified.
5. Confirm dependency vulnerability gating tooling: `govulncheck` (binary-level) and `osv-scanner` (module-level), with severity filters.
6. Determine how to map NC2's two-PATCH-vs-PUT update endpoints to Terraform's diff-driven Update method (FR-010b).
7. Determine how to model `nc2_cloud_account` and `nc2_cloud_account_region` destroy semantics for the no-DELETE case (FR-006, FR-007).
8. Determine the canonical layout for per-resource sensitive registries (FR-002 pattern + per-resource overrides) and the lint mechanism for newly-discovered fields (FR-002a).
9. Determine the OpenAPI coverage report format and gate behavior (FR-021, FR-021a).
10. Verify that `tflog` can emit structured JSON records routable via `TF_LOG` / `TF_LOG_PATH` per FR-020c, and confirm the recommended field naming.

**Phase 0 status**: COMPLETE — no `NEEDS CLARIFICATION` markers remain.

## Phase 1 — Design & Contracts

Prerequisites: `research.md` complete.

Generated artifacts (in this directory):

1. **`data-model.md`** — every entity from the spec (Organization, Cloud Account, Region, Availability Zone, Cluster {AWS|Azure|GCP}, VPC, VNet, SSH Key, Prism Central, Remote Storage Profile, Cloud Resource, Notification, Task), expressed as a Terraform schema (attribute name, type, required/optional/computed, sensitive, plan modifiers, validators, NC2 endpoint mapping, lifecycle states).
2. **`contracts/`** — interface contracts for the Terraform-facing surface:
   - `contracts/README.md` — convention and index.
   - `contracts/provider.json` — full provider config schema (auth, base URL, `ca_bundle`, profile, timeouts).
   - `contracts/resources/nc2_organization.json` — canonical managed resource schema.
   - `contracts/resources/nc2_aws_cluster.json` — most complex managed resource schema.
   - `contracts/datasources/nc2_cloud_account_regions.json` — canonical data source schema.
   - `contracts/actions/nc2_cluster_condemn_host.json` — canonical action schema.
   The remaining 4 managed resources, 17 data sources, and 7 actions follow the same conventions and are enumerated as task items by `/speckit.tasks`; their full JSON schemas are produced during implementation and validated by `tools/coverage-check`.
3. **`quickstart.md`** — getting-started for a developer working on the provider (clone, build, unit-test, acceptance-test, example HCL).
4. **Agent context update** — `.cursor/rules/specify-rules.mdc` updated to reference this plan between the `<!-- SPECKIT START -->` and `<!-- SPECKIT END -->` markers.

### Post-Phase-1 Constitution Check (re-evaluation)

After producing the data model, contracts, and quickstart, re-evaluate the four principles against the more concrete artifacts:

- **Library-First**: data-model.md confirms the natural decomposition into per-entity packages; contracts confirm a stable public surface per library. PASS.
- **TDD**: each contract entry has at least one corresponding test scenario in data-model.md; the OpenAPI coverage tool itself has unit tests scoped in `tools/coverage-check`. PASS.
- **Functional Patterns**: data-model.md attribute mappings are deterministic functions of NC2 responses; the contract files express schema only, no behavior. PASS.
- **Additional constraints**: every library has its own README in the planned layout; sensitive registries are co-located with the resource they cover; no hidden I/O is introduced by the design. PASS.

**Overall Constitution Check (post-Phase-1)**: **PASS** — no new violations introduced by the design phase; no Complexity Tracking entries required.
