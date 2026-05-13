# Quickstart — terraform-provider-nc2

**Audience**: developers working on the provider (not yet end users of the released provider).
**Feature**: Terraform Provider for Nutanix Cloud Clusters (NC2)
**Plan**: [plan.md](./plan.md)
**Date**: 2026-05-13

This document is the developer onramp for the `terraform-provider-nc2` codebase. It complements the spec, plan, research, and data-model documents in this directory by showing how to actually clone, build, test, and exercise the provider against either a fake NC2 server (unit-test path) or a real sandbox NC2 tenant (acceptance-test path).

End-user documentation (provider configuration in HCL, resource reference, examples per resource) will live under `docs/` in the implementation repository and is generated from the schema per FR-029 / FR-030 — that material is not duplicated here.

## 1. Prerequisites

| Tool | Version | Why |
|---|---|---|
| Go | 1.24+ | Toolchain target of `terraform-plugin-framework` 1.x |
| Terraform CLI | 1.7+ | Required by `terraform-plugin-testing` |
| `make` | any | Convenience target driver |
| `git` | any | Source control |
| Optional: a sandbox NC2 tenant with API key + key ID + issuer UUID | – | Acceptance tests only |

For the security-hardened release pipeline (only needed if running `make release` locally):

| Tool | Why |
|---|---|
| `goreleaser` ≥ 2.x | Cross-compile, archive, GPG sign |
| `cosign` ≥ 2.x | Sigstore keyless signatures (FR-032) |
| `slsa-verifier` | Verify SLSA L3 provenance attestation |
| A GPG key registered with the Terraform Registry | GPG-sign SHA256SUMS |

## 2. Clone and bootstrap

```bash
git clone https://github.com/<org>/terraform-provider-nc2.git
cd terraform-provider-nc2

# Install all tools listed in tools.go (kept under //go:build tools)
make tools
```

`make tools` installs `govulncheck`, `osv-scanner`, `tfplugindocs`, `golangci-lint`, and the local `coverage-check` / `sensitive-lint` binaries into `$GOBIN`.

## 3. Build the provider locally

```bash
make build
```

Produces `./terraform-provider-nc2` in the repo root. To make a local Terraform installation use it, drop a `dev_overrides` block into `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/nutanix/nc2" = "/absolute/path/to/repo"
  }
  direct {}
}
```

## 4. Configure provider credentials

Three sources, highest precedence first (per FR-001a):

### 4a. HCL block (test/dev only; do not commit real secrets)

```hcl
provider "nc2" {
  api_key = var.nc2_api_key
  key_id  = var.nc2_key_id
  issuer  = var.nc2_issuer
}
```

### 4b. Environment variables

```bash
export NC2_API_KEY='<your_api_key>'
export NC2_KEY_ID='<your_key_id>'
export NC2_ISSUER='<your_issuer_uuid>'
```

### 4c. Credentials file (default `~/.nc2/credentials`)

```ini
[default]
api_key = <your_api_key>
key_id  = <your_key_id>
issuer  = <your_issuer_uuid>

[ci-sandbox]
api_key = <other_api_key>
key_id  = <other_key_id>
issuer  = <other_issuer_uuid>
```

Select a profile:

```hcl
provider "nc2" {
  profile = "ci-sandbox"
}
```

or `export NC2_PROFILE=ci-sandbox`.

When the same credential attribute is supplied by more than one source AND values differ, the provider emits a warning naming both sources and uses the higher-precedence value (FR-001a). This is testable via `internal/provider/config_test.go::TestCredentialSourceConflict`.

### 4d. Custom CA bundle (corporate proxy or test mirror)

```hcl
provider "nc2" {
  base_url   = "https://nc2-internal.example.test/api/v2"
  ca_bundle  = "/etc/ssl/private/corp-root-ca.pem"
}
```

The PEM file is appended to (not replacing) the OS trust store at startup per FR-003b. **There is no `insecure_skip_verify` option** in any build (FR-003b, FR-003c).

## 5. Run the test suites

### 5a. Unit tests (every PR)

```bash
make test
```

Equivalent to `go test -race ./...`. Runs against fakes (`net/http/httptest`); does not touch a real NC2 endpoint. Coverage gate is checked at ≥ 80 % per SC-005:

```bash
make coverage
```

### 5b. Lint and code quality

```bash
make lint               # golangci-lint with the project's .golangci.yml
make doc-lint           # revive 'exported' rule (SC-007: every exported symbol documented)
```

### 5c. OpenAPI coverage gate (FR-021, FR-021a)

```bash
make openapi-coverage
```

Runs `tools/coverage-check`. Fails the build if any operation in `openapi/openapi.json` is unmapped by a resource / data source / action `OperationMappings` declaration. Output goes to `tools/coverage-check/output/coverage-report.json`.

### 5d. Sensitive-field lint (FR-002a)

```bash
make sensitive-lint               # advisory mode
make sensitive-lint-strict        # CI mode — fails on missing classifications
```

### 5e. Vulnerability scans (FR-032a)

```bash
make vuln
```

Runs `govulncheck` against the compiled binary and `osv-scanner` against `go.mod`/`go.sum`. Fails the build on any HIGH or CRITICAL CVE.

### 5f. Acceptance tests (gated)

```bash
TF_ACC=1 \
NC2_API_KEY=… \
NC2_KEY_ID=… \
NC2_ISSUER=… \
make testacc
```

Acceptance tests provision and tear down **real** NC2 resources and incur cloud spend. They are gated by `TF_ACC=1` and the credential env vars; without them, every acceptance test is `t.Skip()`-ed.

Per FR-023, every managed resource has an acceptance test performing the full lifecycle:

```
terraform init
  → terraform plan      # shows create
  → terraform apply
  → terraform plan      # MUST report "No changes"     (drift-free, SC-002)
  → (if updatable) terraform apply with one field changed
    → terraform plan    # MUST report "No changes"     (drift-free after update)
  → terraform destroy
  → terraform plan      # MUST report no resources     (clean destroy, SC-003)
```

## 6. Minimal end-to-end example (MVP slice — User Story 1)

The smallest configuration that exercises the MVP (US1: organization + cloud account) is:

```hcl
terraform {
  required_providers {
    nc2 = {
      source  = "nutanix/nc2"
      version = "~> 0.1"
    }
  }
}

provider "nc2" {}  # credentials from env vars or ~/.nc2/credentials

resource "nc2_organization" "demo" {
  name        = "speckit-quickstart"
  description = "Demo organization for quickstart"
}

resource "nc2_cloud_account" "demo_aws" {
  organization_id = nc2_organization.demo.id
  cloud_provider  = "aws"
  name            = "demo-aws-account"

  credentials = {
    aws = {
      access_key_id     = var.aws_access_key_id
      secret_access_key = var.aws_secret_access_key  # auto-redacted by FR-002
    }
  }
}
```

Run:

```bash
terraform init
terraform plan    # 2 resources to add
terraform apply
terraform plan    # MUST report "No changes."
terraform destroy
terraform plan    # MUST report 0 resources
```

The second `terraform plan` proving zero drift, and the post-destroy plan proving clean state, are SC-002 and SC-003 in action.

## 7. Adding a new managed resource — the TDD loop

Per Constitution Principle II (NON-NEGOTIABLE), every behavior change is preceded by a failing test in the same diff. The canonical flow for adding a new resource is:

1. Write the failing acceptance test in `tests/acceptance/<resource>_test.go`.
   - Confirm it fails (resource type doesn't exist yet).
2. Write the failing unit tests in `internal/resources/<resource>/resource_test.go` (schema shape, plan modifiers, sensitive registry, drift-free round-trip).
   - Confirm they fail.
3. Write the contract file in `contracts/resources/<resource>.json` (this directory).
   - Re-run `make openapi-coverage` — must show the resource's `OperationMappings` slot the relevant OpenAPI operation IDs.
4. Implement the resource (`resource.go`, `schema.go`, `sensitive.go`, `openapi_mapping.go`, `README.md`).
5. Run `make test`; confirm tests pass.
6. Run `make doc-lint` to confirm every exported symbol is documented.
7. Re-run `make openapi-coverage`, `make sensitive-lint-strict`, and `make vuln`.
8. Update `docs/` (auto-generated where possible) and add a runnable example under `examples/resources/<resource>/`.
9. Submit PR; CI runs all of the above plus a scheduled acceptance test lane.

The "test-required" gate (FR-027) fails the PR if any new exported symbol is missing a test, even if `go test` passes.

## 8. Adding a new action — same loop, narrower surface

Actions follow the same TDD loop. The key difference is that an action has only an `Invoke` lifecycle method (no Read / Update / Delete). The acceptance test asserts:

- The NC2 endpoint is called with the right body.
- For async actions, the task poll completes successfully.
- The action's `result` attribute exposes `task_id`, `status`, `http_status`, `error_code`, `error_message` (FR-017).
- An audit record is emitted (FR-020a..d, asserted via a `tflog` capture helper).

## 9. Where to read next

- [`spec.md`](./spec.md) — what the provider must do, framed in user-visible terms.
- [`plan.md`](./plan.md) — the implementation plan, technical context, project structure, and constitution check.
- [`research.md`](./research.md) — the 10 Phase 0 research questions and their resolved answers.
- [`data-model.md`](./data-model.md) — every entity's schema, NC2 operation mapping, lifecycle, and sensitive registry.
- [`contracts/README.md`](./contracts/README.md) — the contract file convention and the canonical examples committed in this phase.
- [`checklists/requirements.md`](./checklists/requirements.md) — spec quality checklist (passed) and the revision log.
- `openapi/openapi.json` (repository root) — the authoritative API surface (49 operations across 40 paths).
- `.specify/memory/constitution.md` — the project constitution (v1.0.0): Library-First, TDD (non-negotiable), Functional patterns.

## 10. After implementation: end-user quickstart

Once the provider is built and published, the end-user-facing quickstart will live at `docs/guides/getting-started.md` in the implementation repo. The Spec Kit quickstart (this file) is for contributors.
