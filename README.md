# terraform-provider-nc2

A Terraform provider for **Nutanix Cloud Clusters (NC2)**, covering 100% of the
NC2 v2 API surface — 6 managed resources, 18 data sources, and 8 actions — with
strict drift-free / clean-destroy quality bars on every managed resource and a
security-forward release pipeline (GPG + Sigstore + SLSA Build L3 provenance).

> **Status**: this repository is in active development. The specification, plan,
> contracts, and task list are committed; implementation is in progress. See
> [`specs/001-nutanix-nc2-provider/`](specs/001-nutanix-nc2-provider/) for the
> authoritative design documents.

## Quick links

| Want to … | Read |
|---|---|
| Understand what the provider does | [`spec.md`](specs/001-nutanix-nc2-provider/spec.md) |
| Understand how it's built | [`plan.md`](specs/001-nutanix-nc2-provider/plan.md) |
| See the entity → Terraform schema map | [`data-model.md`](specs/001-nutanix-nc2-provider/data-model.md) |
| See the Phase 0 research decisions | [`research.md`](specs/001-nutanix-nc2-provider/research.md) |
| See the dependency-ordered task list | [`tasks.md`](specs/001-nutanix-nc2-provider/tasks.md) |
| Build / test the provider locally | [`quickstart.md`](specs/001-nutanix-nc2-provider/quickstart.md) |
| See the project constitution | [`.specify/memory/constitution.md`](.specify/memory/constitution.md) |
| Browse the NC2 API surface | [`openapi/openapi.json`](openapi/openapi.json) |

## Project principles (constitution v1.0.0, NON-NEGOTIABLE where marked)

- **I. Library-First** (non-negotiable): every feature ships as a self-contained,
  independently testable library before any caller depends on it.
- **II. Test-Driven Development** (non-negotiable): the Red → Green → Refactor
  cycle is mandatory for every behavior change. CI blocks any PR that introduces
  a new exported symbol without a corresponding test.
- **III. Functional Programming Patterns**: pure functions by default; mutable
  shared state isolated to clearly-marked integration edges.

## Security posture (one-line summary)

| | |
|---|---|
| TLS | Strict verification, no skip-verify in any build (FR-003b/c). |
| Credentials | Three sources, precedence block > env > credentials file; no external-secret-store imports (FR-001a, FR-001b). |
| Sensitive fields | Pattern + per-resource registry; lint catches new fields (FR-002, FR-002a). |
| Audit log | One structured JSON record per NC2 API call via `tflog` (FR-020a..d). |
| Release | GPG + Sigstore + SLSA L3, gated by `govulncheck` + `osv-scanner` (FR-032, FR-032a). |

Detailed security guide: [`docs/guides/security-hardening.md`](docs/guides/security-hardening.md)
*(produced in Phase 8; placeholder for now)*.

## Quick build (developer)

```bash
git config core.hooksPath .githooks   # one-time: enable the pre-commit hook
make tools                            # install Go-installable CI tools
make test                             # run unit tests with -race
make build                            # produce ./terraform-provider-nc2
make lint                             # run golangci-lint
```

Full developer onramp, including local `terraform { dev_overrides {} }` setup
and acceptance-test instructions, is in
[`specs/001-nutanix-nc2-provider/quickstart.md`](specs/001-nutanix-nc2-provider/quickstart.md).

## Contributing

Contributions must follow the constitution and the PR checklist in
[`.github/PULL_REQUEST_TEMPLATE.md`](.github/PULL_REQUEST_TEMPLATE.md). In
particular:

- Every behavior change ships with a failing-then-passing test in the same diff.
- Every internal library has its own `README.md` and tests.
- No `InsecureSkipVerify`, no external secret store imports.
- Every new exported Go symbol has a doc comment.

## License

This project is licensed under the [Mozilla Public License 2.0](LICENSE), the
established convention for Terraform-ecosystem providers.
