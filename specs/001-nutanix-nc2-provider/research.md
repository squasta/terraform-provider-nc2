# Phase 0 — Research Findings

**Feature**: Terraform Provider for Nutanix Cloud Clusters (NC2)
**Plan**: [plan.md](./plan.md)
**Date**: 2026-05-13

This document records the Phase 0 research outputs. Every entry is structured as **Decision / Rationale / Alternatives considered**, per the Spec Kit plan workflow. No `NEEDS CLARIFICATION` markers remain after this phase.

---

## R-01: Terraform Plugin Framework version and Action support

**Decision**: Build on `github.com/hashicorp/terraform-plugin-framework` v1.x (current GA), pinned to the latest published `v1.<latest>.<latest>` at first commit and updated via Dependabot. Use the framework's `action` package for the 8 non-CRUD operational endpoints (FR-016). Pin Go to 1.24+ to match the framework's supported toolchain.

**Rationale**:

- The user's explicit instruction ("latest terraform provider framework") rules out the legacy `terraform-plugin-sdk/v2`. Modern plugin development on terraform-plugin-framework is the supported path.
- The framework's `action` concept (alongside resources, data sources, and ephemeral resources) is the natural fit for imperative one-shot endpoints in NC2 (condemn host, support tunnel control, scale/upgrade Flow Gateway, start recovery, acknowledge notification). Modeling these as actions avoids the anti-pattern of "managed resources with no read" and avoids state-corruption risks. This is what the spec already assumes (FR-016, FR-016a).
- The framework provides the supporting test harness (`terraform-plugin-testing`) and structured logging (`tflog`) we need for FR-020a..d and FR-023..FR-026.

**Caveat captured during research**: If the `action` API is still flagged as `preview`/`experimental` in the framework release we pin to at implementation time, we will (a) consume it via the explicit preview import path, (b) record the framework version in `CHANGELOG.md` and the provider README, and (c) plan for a major-version bump of the provider when the framework promotes actions to GA if any breaking changes occur. This is recorded as a versioning risk in the Release section of the plan, not a blocker — the design is correct regardless.

**Alternatives considered**:

- `terraform-plugin-sdk/v2`: legacy. Excluded by user instruction and by framework feature gap (no actions, weaker plan-modifier semantics).
- Custom RPC implementation: rejected — re-implementing the Terraform plugin protocol is an enormous undertaking and gains nothing.
- Modeling actions as managed resources with no-op read: rejected — violates the spec (FR-016a) and produces state-corruption risk.

---

## R-02: Exact JWT signing recipe required by NC2

**Decision**: Implement JWT minting per the recipe in `openapi/openapi.json`'s `info.description` and the NC2 quick-start sample:

1. Build the JWT payload with:
   - `aud = "https://apikeys.nutanix.com"`
   - `iat = now()`
   - `exp = iat + 300s` (5 minutes; Nutanix's recommended maximum)
   - `iss = <issuer UUID from FR-001a-resolved credentials>`
   - `metadata = {"reason": "terraform-provider-nc2"}`
   - `context = {}`
2. Compute the JWT signing secret as `base64_standard(HMAC_SHA512(api_key, key_id))`. Note the unusual double-step: the input bytes to HMAC are the `key_id` itself, and the **base64 of the HMAC output** is then used as the secret passed to the JWT encoder.
3. Encode the JWT with algorithm `HS512` and header `kid = <key_id>`. The library used is `github.com/golang-jwt/jwt/v5`.

Sign and emit the resulting compact-serialized JWT in the HTTP `Authorization: Bearer <jwt>` header on every NC2 API call.

**Rationale**:

- The recipe is documented in two places in the source material (the OpenAPI `info.description` and the public reference page), so we have high confidence.
- The unusual "base64-encoded HMAC as the JWT secret" pattern is non-obvious and will be the most likely source of misconfiguration; encoding the recipe as a pure function (`auth.MintJWT(creds Credentials, now time.Time) (Token, error)`) with table-driven tests against known-good fixtures protects against regression.

**Alternatives considered**:

- `crypto/hmac` + custom JWT serialization: rejected — needlessly re-implements the JWT compact-serialization spec; introduces bugs.
- `square/go-jose` library: viable but heavier-weight; `golang-jwt/jwt/v5` is the established Terraform-ecosystem choice and has a smaller attack surface.
- Single-step HMAC (signing the JWT directly with `api_key`): rejected — contradicts the documented Nutanix recipe; would fail authentication at the server.

---

## R-03: OpenAPI parser for the CI-only tools

**Decision**: Use `github.com/getkin/kin-openapi` for the CI binaries `tools/coverage-check` and `tools/sensitive-lint`. Import behind a `// +build tools` constraint so the dependency does **not** appear in the provider's runtime binary.

**Rationale**:

- `kin-openapi` is the most active and best-maintained OpenAPI 3.0 parser in the Go ecosystem; it cleanly parses `openapi/openapi.json` (verified: 49 operations across 40 paths, 13 tags loaded without error).
- The tools-only-build-tag pattern is a well-established Go convention (HashiCorp uses it for `gofumpt`, `tfplugindocs`, etc.) — it keeps `go.mod` honest while preventing the dependency from linking into the production binary.

**Alternatives considered**:

- `pb33f/libopenapi`: newer, very capable, but younger ecosystem and adds a Go-generics requirement that complicates the tools dependency graph. Acceptable second choice if `kin-openapi` regresses.
- Hand-rolled JSON walker: rejected — `openapi/openapi.json` is large (8 390 lines, 408 KB) and uses every OpenAPI 3.0 feature including `$ref` chasing; rolling our own is high-risk and low-value.

---

## R-04: Cross-compile, signing, and provenance tooling

**Decision**: Use `goreleaser/goreleaser` to drive cross-compilation, `aevea/gh-action-cosign` (or direct `sigstore/cosign-installer`) for keyless Sigstore signatures, `slsa-framework/slsa-github-generator` for SLSA Build Level 3 provenance, and a long-lived GPG key (stored as a GitHub Actions secret) for Terraform Registry compatibility. Wire all four into a single `release.yml` GitHub Actions workflow triggered by `v*.*.*` tags.

**Rationale**:

- GoReleaser is the de-facto Terraform Registry release tool; HashiCorp's published guides recommend it.
- `slsa-github-generator` produces an in-toto attestation that satisfies SLSA Build L3 when the workflow runs on a public GitHub repository with restricted secrets — exactly our target. The attestation is published as a release asset alongside the GPG-signed SHA256SUMS file.
- Sigstore cosign keyless signing leverages GitHub Actions' OIDC token; no long-lived signing key to rotate. Per-artifact `.sig` and `.pem` accompany every binary.
- GPG remains necessary because the Terraform Registry's current consumer protocol verifies GPG signatures on SHA256SUMS. Until/unless the Registry migrates to Sigstore-native verification, we ship both.

**Alternatives considered**:

- Manual `go build` + `tar`/`zip` + manual GPG: rejected — error-prone, not reproducible, and incompatible with SLSA L3 (which requires an unmodified hermetic build).
- `slsa-verifier` for self-verification only (no provenance generation): rejected — we want third parties to be able to verify provenance independently.
- Replacing GPG with Sigstore entirely: deferred — possible in a future release when the Terraform Registry universally supports Sigstore.

---

## R-05: Dependency vulnerability scanning

**Decision**: Run two complementary scans on every release, blocking on HIGH or CRITICAL severity (FR-032a):

1. `golang.org/x/vuln/cmd/govulncheck` against the compiled binary — catches issues in transitive dependencies actually reachable from the binary's call graph.
2. `github.com/google/osv-scanner` against `go.mod`/`go.sum` — catches issues in any declared dependency, including those not yet reachable.

Severity filtering uses each tool's native CVSS-derived classification. MEDIUM and LOW findings are written to `vulnerability-report.json` (published as a release asset per FR-032b) and to `CHANGELOG.md`, but do not block the release.

**Rationale**:

- Two scanners cover complementary attack surfaces: `govulncheck` is precise (reachability-aware, fewer false positives but may miss issues in unreachable code), `osv-scanner` is exhaustive (catches everything declared but noisier). Running both gives both signal and coverage.
- Blocking on HIGH/CRITICAL is the conventional choice; LOW/MEDIUM in transitive deps is usually waitable, especially for a Go binary where most code is statically linked.

**Alternatives considered**:

- Single scanner (`govulncheck` only or `osv-scanner` only): rejected — each catches issues the other misses.
- Third-party SaaS (Snyk, GitHub Dependabot security): rejected as the primary gate — we want CI-gated, offline-runnable checks; these can supplement.

---

## R-06: PATCH vs PUT canonical mapping for updates

**Decision**: For each NC2 resource exposing both `PATCH` and `PUT` update variants (clusters, organizations, cloud accounts, notifications), the provider's Terraform `Update` method canonically uses **`PATCH`** with a body containing only the diff-changed attributes. `PUT` is implemented in the `internal/client` library but is **only** called by an explicit `ReplaceAll()` helper that the provider never invokes from Update. This satisfies FR-010b.

**Rationale**:

- `PATCH` semantically matches Terraform's diff-driven model: Terraform tells the provider which fields changed, and we forward only those to NC2. Using `PUT` for a single-field change would require either re-reading the current state and merging in the diff (expensive, racy) or sending a full snapshot of state including computed fields (state-leakage risk).
- Reserving `PUT` for a documented future "full replacement" code path keeps the option available for compliance/conformance scenarios without polluting normal flows.

**Alternatives considered**:

- Always-`PUT`: rejected — see racy-merge problem above.
- Sometimes-`PATCH`, sometimes-`PUT` based on changed-attribute count: rejected — non-deterministic from the user's perspective; very hard to test exhaustively.

---

## R-07: Destroy semantics for `nc2_cloud_account` and `nc2_cloud_account_region`

**Decision**: For both `nc2_cloud_account` (no DELETE endpoint) and `nc2_cloud_account_region` (no remove endpoint), implement Terraform's `Delete` method as:

1. Emit a non-fatal diagnostic at plan time (during `PlanResourceChange`) with severity `Warning`, content:
   > "NC2 v2 API exposes no DELETE endpoint for this resource. `terraform destroy` will remove the resource from Terraform state only; the underlying NC2 entity must be removed manually through the NC2 console. See the resource documentation for details."
2. In `Delete`, perform **no** NC2 API call; return success. The plugin framework removes the resource from state by virtue of `Delete` returning no error.
3. Add an idempotency guard in `Read`: if a stale state entry references an NC2 entity that no longer exists (`404`), the entry is removed from state per FR-009 — the normal out-of-band-deletion path.

This satisfies FR-006 and FR-007.

**Rationale**:

- This is the safest and most honest behavior given the upstream API constraint. It avoids silent failures (no false-success destroy that leaves NC2 entities orphaned) and avoids fabricated errors (we don't try to call a non-existent endpoint).
- The plan-time warning makes the limitation visible *before* the user clicks "yes", which is the right time to surface it.
- Documenting the behavior in the resource reference page (per FR-006's "MUST be documented in the resource reference page") closes the loop.

**Alternatives considered**:

- Fail-on-destroy: rejected — would leave users unable to clean up Terraform state without resource manipulation, which is worse than the warned no-op.
- Silent state-only removal with no warning: rejected — user would think NC2 entity was deleted; bad experience.
- Treat as immutable (no Delete at all): rejected — Terraform requires `Delete` to be implementable; not having one breaks `terraform destroy` for the whole workspace.

---

## R-08: Per-resource sensitive registry layout and the FR-002a lint mechanism

**Decision**: Each resource / data source / action package exports a `Sensitive()` function returning a deterministic `[]string` of attribute paths the package classifies as sensitive in addition to the FR-002 pattern-list defaults. Attribute paths use the dotted JSON-pointer-ish form (`credentials.access_key.value`, `network.prism_element_access_policy.ip_addresses`). The provider's `redact` library reads both the FR-002 pattern list and these per-package registries at startup.

For FR-002a (lint), `tools/sensitive-lint` walks `openapi/openapi.json` and computes, for each managed resource and data source:

- The set of attribute paths whose JSON name matches an FR-002 pattern (case-insensitive substring match).
- The set of attribute paths in the per-package `Sensitive()` registry.

The tool fails CI in `strict` mode if a path matched by an FR-002 pattern is **not** included in the package's registry (auto-mark), or if a path is in the registry but the OpenAPI no longer references it (drift). MEDIUM-severity warnings are emitted for newly-discovered attribute paths that do **not** match a pattern and are **not** in a registry — these prompt the maintainer to classify.

**Rationale**:

- Co-locating the registry with the resource keeps the source of truth next to the schema, matching the constitution's "co-located documentation" principle.
- A pure function returning the registry is easy to test (table-driven assertions against fixtures), aligns with Principle III, and is trivially callable from `tools/sensitive-lint` via Go reflection over the package's exported function.
- The strict/non-strict modes let local development run the lint in advisory mode while CI runs it in blocking mode.

**Alternatives considered**:

- Centralized registry in `internal/redact`: rejected — splits the source of truth away from the resource it covers; violates Library-First.
- Tag-based registry (Go struct tags on schema types): viable but inflexible for nested attributes; rejected as primary mechanism but may be added as a supplement later.

---

## R-09: OpenAPI coverage report format and gate behavior

**Decision**: `tools/coverage-check` produces a JSON report (`coverage-report.json`) at `tools/coverage-check/output/coverage-report.json` and a human-readable Markdown summary. Schema:

```json
{
  "openapi_source": "openapi/openapi.json",
  "openapi_version": "v2",
  "total_operations": 49,
  "covered": 49,
  "uncovered": [],
  "mapping": [
    {
      "operation_id": "CPanelWeb.Api.ClusterController.create_aws",
      "method": "POST",
      "path": "/clusters/aws",
      "tag": "clusters",
      "mapped_to": {
        "kind": "resource",
        "type_name": "nc2_aws_cluster",
        "operation": "Create"
      }
    }
  ]
}
```

The tool exits non-zero if `uncovered` is non-empty. CI uploads `coverage-report.json` as a build artifact on every PR so reviewers can see the mapping.

Mapping is sourced from a Go file in each resource / data source / action package — for example, `internal/resources/aws_cluster/openapi_mapping.go` declares:

```go
var OperationMappings = []openapi.Mapping{
    {OperationID: "CPanelWeb.Api.ClusterController.create_aws", TerraformOp: "nc2_aws_cluster.Create"},
    {OperationID: "CPanelWeb.Api.ClusterController.show",        TerraformOp: "nc2_aws_cluster.Read"},
    {OperationID: "CPanelWeb.Api.ClusterController.update (2)",  TerraformOp: "nc2_aws_cluster.Update.patch"},
    {OperationID: "CPanelWeb.Api.ClusterController.terminate",   TerraformOp: "nc2_aws_cluster.Delete"},
    {OperationID: "CPanelWeb.Api.ClusterController.update_capacity",        TerraformOp: "nc2_aws_cluster.Update.capacity"},
    {OperationID: "CPanelWeb.Api.ClusterController.update_ssh_key",         TerraformOp: "nc2_aws_cluster.Update.ssh_key"},
    {OperationID: "CPanelWeb.Api.ClusterController.update_license",         TerraformOp: "nc2_aws_cluster.Update.license"},
    {OperationID: "CPanelWeb.Api.ClusterController.update_resource_tags",   TerraformOp: "nc2_aws_cluster.Update.resource_tags"},
    {OperationID: "CPanelWeb.Api.ClusterController.update_access_policy",   TerraformOp: "nc2_aws_cluster.Update.access_policy"},
    {OperationID: "CPanelWeb.Api.ClusterController.hibernate",   TerraformOp: "nc2_aws_cluster.Update.desired_state.hibernate"},
    {OperationID: "CPanelWeb.Api.ClusterController.resume",      TerraformOp: "nc2_aws_cluster.Update.desired_state.resume"},
}
```

`tools/coverage-check` discovers these by scanning `internal/resources/*/openapi_mapping.go`, `internal/datasources/*/openapi_mapping.go`, `internal/actions/*/openapi_mapping.go` at build time. New OpenAPI operations not present in any mapping cause the gate to fail (FR-021a).

**Rationale**:

- Declarative mapping files keep the truth source in each package; a missing mapping in the package is exactly the right place to demand a fix.
- JSON output is consumable by humans, CI, and external dashboards.
- The gate is fail-closed (uncovered ⇒ red build), which is the safe default for a coverage claim.

**Alternatives considered**:

- Comment-based annotations parsed from `*.go` files: rejected — fragile, no compile-time validation.
- A single central `coverage.yaml` file: rejected — violates Library-First, gets out of sync.

---

## R-10: `tflog` structured audit emission

**Decision**: Use `terraform-plugin-log/tflog.SetField` + `tflog.Info` to emit one structured record per NC2 HTTP call. The framework already JSON-serializes the per-call field map when `TF_LOG` is set to JSON mode (which it is by default in the framework's log mode). Field names match FR-020a exactly. Subsystem name: `audit` (so users can filter via `TF_LOG_SDK_HELPER_RESOURCE=OFF TF_LOG_PROVIDER_AUDIT=INFO`).

The emitter lives in `internal/audit/audit.go` as `func Record(ctx context.Context, r AuditRecord)` — a pure-by-construction function (it returns no error, and its only side effect is the `tflog.Info` call which is itself a thin wrapper over the framework's log channel). The audit record type is a flat struct:

```go
type AuditRecord struct {
    CorrelationID string
    TerraformOp   string
    HTTPMethod    string
    Path          string
    Status        int
    LatencyMs     int64
    NC2TaskID     string  // empty if absent
    NC2ErrorCode  string  // empty if absent
}
```

The `client.Do` function constructs the record and invokes `audit.Record` before returning, regardless of success/failure.

**Rationale**:

- `tflog` is already a required dependency (FR-020 error surfacing uses it), so we add no new dependency.
- Routing via `TF_LOG` / `TF_LOG_PATH` (FR-020c) means we inherit the existing user-facing logging contract; operators don't have to learn a new mechanism.
- The audit record type and its emitter are testable in isolation: a unit test passes a fake `tflog`-recording context and asserts the rendered JSON contains the expected fields with sensitive values redacted.

**Alternatives considered**:

- Writing audit records directly to `os.Stderr` with a custom JSON encoder: rejected — bypasses the framework's logging stack and breaks `TF_LOG_PATH` routing.
- Using OpenTelemetry traces: viable but heavyweight; deferred to a future enhancement if user demand emerges. The current `tflog` design does not preclude later addition of an OTel exporter.
- Emitting at DEBUG instead of INFO: rejected — INFO is the right level because the audit log is a deliberate operator-facing artifact, not a debug aid.

---

## Phase 0 status

All 10 research questions are resolved. No `NEEDS CLARIFICATION` markers remain. Proceed to Phase 1.
