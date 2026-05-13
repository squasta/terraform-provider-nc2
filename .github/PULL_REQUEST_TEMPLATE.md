<!-- Per task T008. Encodes the Constitution Check (`.specify/memory/constitution.md`)
     as a per-PR checklist. Reviewers MUST verify each box before merging. -->

## Summary

<!-- 1–3 sentence summary of the change. Reference the user story (US1..US5) and
     task IDs (T0xx) from `specs/001-nutanix-nc2-provider/tasks.md` if applicable. -->

## Constitution Check

The project constitution is at `.specify/memory/constitution.md` (v1.0.0). All three principles are **NON-NEGOTIABLE**. Tick every box that applies to this PR.

### Principle I — Library-First

- [ ] Every new feature ships as a self-contained, independently testable library before any caller depends on it.
- [ ] Each new library has a single, clearly-stated purpose (no "organizational" / "junk drawer" libraries).
- [ ] Each new library has its own `README.md` documenting purpose, public API, inputs, outputs, and errors.
- [ ] No undocumented public APIs are introduced.

### Principle II — Test-Driven Development (NON-NEGOTIABLE)

- [ ] Every behavior change ships with a corresponding test in the same diff.
- [ ] Each test was first verified to FAIL for the right reason before the implementation was written.
- [ ] The `*_test.go` file lives alongside the package it tests.
- [ ] No skipped tests without an `//nolint:tskip` annotation that includes a tracking issue.
- [ ] Test names follow the pattern `Test<Subject>_<Scenario>_<ExpectedOutcome>`.

### Principle III — Functional Programming Patterns

- [ ] New functions are pure where practical (no side effects, deterministic on inputs).
- [ ] Mutable shared state is avoided, isolated, or made explicit.
- [ ] Composition is used over inheritance / embedding-for-behavior-reuse.
- [ ] Data is treated as immutable by default; mutations are surfaced explicitly.
- [ ] Imperative / OO patterns are confined to integration edges (HTTP client, JWT cache, logger).

### Additional Cross-Cutting Constraints

- [ ] No `InsecureSkipVerify` introduced anywhere (FR-003b).
- [ ] No imports of external-secret-store client libraries (FR-001b).
- [ ] All new attributes matching FR-002 patterns are classified in a per-resource sensitive registry (FR-002a).
- [ ] Every new public Go symbol has a doc comment (SC-007).
- [ ] Every new managed resource / data source / action has a docs page and a runnable example (FR-029, SC-006).
- [ ] The OpenAPI coverage report (`make openapi-coverage`) still passes (FR-021, FR-021a).
- [ ] `make sensitive-lint-strict` passes (FR-002a).
- [ ] `make vuln` reports no HIGH or CRITICAL CVEs (FR-032a).

## Test Plan

<!-- List the scenarios you actually verified. Be specific: include commands run,
     unit / acceptance test names, fixtures used. The auto-CI gates do not relieve
     you from describing what *you* checked.
-->

- [ ] `make test` ........................... unit tests pass with `-race`
- [ ] `make coverage` ....................... line coverage ≥ 80%
- [ ] `make lint` + `make doc-lint` ......... clean
- [ ] `make openapi-coverage` ............... `coverage-report.json` shows the new operation mapped
- [ ] `make sensitive-lint-strict` .......... clean
- [ ] `make vuln` ........................... clean
- [ ] Acceptance test added (managed resource / data source / action changes)
- [ ] Acceptance test exercised against a sandbox NC2 tenant (`TF_ACC=1 make testacc`)

## Spec References

<!-- Link the FRs / SCs / User Stories that this PR delivers or advances. -->

- Functional requirements: <!-- e.g. FR-001a, FR-020a -->
- Success criteria: <!-- e.g. SC-002, SC-005 -->
- Task IDs: <!-- e.g. T046, T053 -->

## Breaking Changes

- [ ] No breaking changes
- [ ] Breaking change documented in `CHANGELOG.md` with migration notes (FR-031)
