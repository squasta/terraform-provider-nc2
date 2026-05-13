<!--
SYNC IMPACT REPORT
==================
Version change: (uninitialized template) → 1.0.0
Bump rationale: First ratification of a populated constitution. No prior versioned
content existed (all placeholder tokens); per semantic versioning policy this is a
1.0.0 initial ratification rather than a MAJOR bump from a prior version.

Principle inventory (template had 5 placeholders; user-specified set has 3):
  - [PRINCIPLE_1_NAME]   → I. Library-First (NON-NEGOTIABLE)
  - [PRINCIPLE_2_NAME]   → II. Test-Driven Development (NON-NEGOTIABLE)
  - [PRINCIPLE_3_NAME]   → III. Functional Programming Patterns
  - [PRINCIPLE_4_NAME]   → (removed; user specified only 3 principles)
  - [PRINCIPLE_5_NAME]   → (removed; user specified only 3 principles)

Added sections:
  - Additional Constraints (concrete library + purity rules)
  - Development Workflow (review gates, CI gates, refactor policy)
  - Governance (amendment + compliance procedure)

Removed sections:
  - [SECTION_2_NAME] / [SECTION_3_NAME] placeholders (replaced by named sections above)

Templates and docs reviewed for propagation:
  ✅ .specify/templates/plan-template.md         (Constitution Check section is
                                                  generic — "Gates determined based on
                                                  constitution file" — no edit required;
                                                  gates will be derived per-feature)
  ✅ .specify/templates/spec-template.md         (no constitution-specific content;
                                                  remains compatible)
  ✅ .specify/templates/tasks-template.md        (UPDATED — test tasks reclassified
                                                  from OPTIONAL to MANDATORY to match
                                                  Principle II — TDD non-negotiable)
  ✅ .specify/templates/constitution-template.md (template form unchanged; only the
                                                  populated copy in .specify/memory/
                                                  is updated)
  ⚠ .specify/templates/commands/*.md            (directory does not exist in this
                                                  project; nothing to update)
  ⚠ README.md (project root)                    (does not exist; nothing to update)
  ⚠ docs/quickstart.md                          (does not exist; nothing to update)

Follow-up TODOs: none
-->

# Test Spec Kit Constitution

## Core Principles

### I. Library-First (NON-NEGOTIABLE)

Every feature MUST begin life as a standalone library before any application, CLI, service,
or UI layer consumes it. Libraries MUST be self-contained, independently installable, and
independently testable without the surrounding application context. Each library MUST have a
single, clearly stated purpose; "organizational" or "junk drawer" libraries that exist only
to group unrelated code are prohibited. The public API of each library MUST be documented
(purpose, inputs, outputs, errors) at the time the library is introduced — undocumented
public APIs MUST NOT be merged.

**Rationale**: Library-first design forces explicit boundaries, makes reuse and substitution
trivial, and keeps applications thin orchestration layers. It also makes testing inexpensive
because each unit is decoupled from runtime concerns, which directly enables Principle II.

### II. Test-Driven Development (NON-NEGOTIABLE)

TDD is mandatory and strictly enforced. The Red → Green → Refactor cycle MUST be followed
for every behavior change:

1. Write a failing test that captures the desired behavior.
2. Confirm the test fails for the right reason (compile / assertion / runtime) before
   writing any implementation.
3. Write the minimum production code required to make the test pass.
4. Refactor only while the test suite is fully green.

Production code MUST NOT be added or modified without a corresponding test added or modified
first in the same change set. Pull requests that introduce or change behavior without a
failing-then-passing test in the diff MUST be rejected. Test suites MUST be runnable locally
and in CI on every commit; broken or skipped tests MUST be fixed or explicitly removed (with
documented justification) before merging — long-lived skipped tests are prohibited.

**Rationale**: TDD keeps design honest, prevents regression debt, and makes the test suite
an executable specification. Strict enforcement is required because TDD compromised
"sometimes" is TDD never.

### III. Functional Programming Patterns

Functional patterns are the default style. Functions MUST be pure where practical: the same
input MUST produce the same output, with no observable side effects. Mutable shared state
MUST be avoided; when state is required it MUST be isolated, explicit, and small.
Composition (function composition, pipelines, higher-order functions) MUST be preferred
over deep inheritance hierarchies. Data structures SHOULD be immutable by default; in-place
mutation MUST be justified by a measurable need (e.g., a documented performance constraint)
and contained behind a pure interface.

Imperative or object-oriented code is permitted only at integration boundaries (I/O,
runtime adapters, third-party APIs) and MUST be kept thin enough that the core logic
remains pure and unit-testable without mocking control flow.

**Rationale**: Pure, composable code is easier to test (Principle II) and easier to package
as reusable units (Principle I). The three principles reinforce each other; weakening any
one of them weakens the others.

## Additional Constraints

- **Explicit library surface**: Each library MUST publish its public surface explicitly
  (e.g., an `index` / `__init__` / equivalent export manifest). Anything not exported is
  internal and MUST NOT be relied upon by consumers.
- **No hidden I/O in pure modules**: Modules that present themselves as pure (logic, data
  transforms, computations) MUST NOT perform network, filesystem, environment, clock, or
  random reads. Effects MUST be injected from the edge.
- **Co-located documentation**: Each library MUST ship a short README covering purpose,
  install, one usage example, and the public API surface. Missing README blocks merge.
- **Tests live with the library**: Each library's tests MUST live alongside the library
  (not only in an aggregated test suite) so the library remains independently testable per
  Principle I.

## Development Workflow

- **Branching**: Feature work happens on feature branches; merges into the default branch
  require passing CI (tests + lint + format checks).
- **Code review gates** — reviewers MUST verify, before approving:
  1. New or changed behavior is covered by a test that demonstrably failed before
     implementation (Principle II).
  2. New functionality is delivered as, or within, a library with a single clear purpose
     and a documented public surface (Principle I).
  3. New code follows functional patterns; any new mutable shared state, class-based
     abstraction, or hidden side effect is explicitly justified in the PR description
     (Principle III).
- **CI quality gates**: The full test suite MUST be green. Coverage of new lines SHOULD NOT
  decrease overall project coverage. Linters configured to flag non-functional patterns
  (e.g., unused mutation, side effects in modules declared pure) MUST run on every PR.
- **Refactor policy**: A refactor PR MUST NOT change behavior and MUST leave existing tests
  unchanged. If a test must change, the PR is a behavior change and the full TDD cycle
  applies.

## Governance

This constitution supersedes ad-hoc team preferences and prior conventions. Where any
document, template, or process conflicts with this constitution, the constitution wins
until amended.

**Amendments**:

- Amendments MUST be proposed by a pull request that updates this file and any dependent
  templates (`plan-template.md`, `spec-template.md`, `tasks-template.md`, and any command
  or skill files that reference the affected principles).
- Each amendment PR MUST include: rationale, a semantic version bump (MAJOR for principle
  removal or backward-incompatible redefinition, MINOR for new principles or materially
  expanded guidance, PATCH for clarifications and wording fixes), and a refreshed Sync
  Impact Report at the top of this file.
- Amendments MUST be approved by at least one maintainer who is not the author.

**Compliance review**: Every PR's review checklist MUST include a "Constitution Check"
step. Violations MUST either be fixed before merge or recorded as exceptions in the
affected feature's `plan.md` under "Complexity Tracking" with explicit justification and a
remediation plan.

**Runtime guidance**: For agent-driven workflows, refer to the Spec Kit skills under
`.cursor/skills/` and the command extensions under `.specify/extensions/`; these MUST be
kept consistent with the principles above.

**Version**: 1.0.0 | **Ratified**: 2026-05-13 | **Last Amended**: 2026-05-13
