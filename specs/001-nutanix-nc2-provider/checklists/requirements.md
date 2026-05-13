# Specification Quality Checklist: Terraform Provider for Nutanix Cloud Clusters (NC2)

**Purpose**: Validate specification completeness and quality before proceeding to planning
**Created**: 2026-05-13
**Feature**: [spec.md](../spec.md)

## Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

> Note on "No implementation details": the deliverable itself is a Terraform provider, so "Terraform provider", "Terraform Registry", and `terraform plan/apply/destroy` semantics are part of the product surface, not internal implementation choices. Internal implementation choices (Go version, framework version, release tooling) are captured in the Assumptions section rather than in functional requirements, per the spec template's intent.

## Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Success criteria are technology-agnostic (no implementation details)
- [x] All acceptance scenarios are defined
- [x] Edge cases are identified
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Feature Readiness

- [x] All functional requirements have clear acceptance criteria
- [x] User scenarios cover primary flows
- [x] Feature meets measurable outcomes defined in Success Criteria
- [x] No implementation details leak into specification

## Notes

- Items marked incomplete require spec updates before `/speckit.clarify` or `/speckit.plan`.
- Assumptions intentionally absorb decisions the user already implicitly made
  (latest framework, Go, all three clouds, registry distribution), so these did
  not need to remain as open clarifications.
- If the user later disagrees with any assumption (e.g., wants AWS only for v1,
  or wants hibernate modeled as separate action resources), run `/speckit.clarify`
  to override before `/speckit.plan`.

## Revision log

- **2026-05-13 (initial)**: Spec drafted from the public HTML reference at
  `nutanix.dev/api_reference/apis/nc2.html`.
- **2026-05-13 (revision 1)**: Spec re-anchored to the local authoritative
  `openapi/openapi.json` (OpenAPI 3.0.0). The total operation count is now
  concrete (49 operations across 40 paths, 13 tag groups). Targeted updates
  fold in operational facts that the OpenAPI makes precise: JWT specifics
  (HS512, 5-min, MyNutanix audience), the `202 Accepted` + task polling
  contract, the AWS-only access policy endpoint, the absence of a DELETE for
  cloud accounts and a remove-region endpoint (with destroy-semantics caveats),
  the PATCH-vs-PUT canonical mapping, and the demotion of notifications from
  "managed resource" to data source + action. Checklist items are re-verified
  and all still pass.
- **2026-05-13 (revision 2)**: User-story priorities reordered per explicit
  user direction ("User Story 5 must be done first and cluster for all clouds
  after"). The tenant-hierarchy work (organizations + cloud accounts, formerly
  US5 at P3) is now US1 at P1 / MVP. The three former cluster stories
  (old US1 AWS-only, old US2 in-place updates, old US3 Azure+GCP parity)
  are merged into a single US2 at P2 covering full CRUD across all three
  clouds together. Inventory data sources (old US4) become US3, hibernate/
  resume (old US6) becomes US4, and non-CRUD actions (old US7) become US5.
  SC-004 was generalized from "provision a working AWS cluster" to a two-step
  journey covering org/account bootstrap then any-cloud cluster provisioning,
  reflecting the new MVP slicing. Functional Requirements (FR-001..FR-032
  and their sub-FRs) are unchanged: priority reordering is a delivery-order
  decision, not a scope change. Checklist items are re-verified and all
  still pass.
- **2026-05-13 (revision 3 — `/speckit.clarify`, security focus)**: Ran
  the clarification protocol with a security lens. Five questions asked
  and answered; a `## Clarifications` section was added to the spec under
  session heading `Session 2026-05-13` recording each Q/A. Functional
  Requirements expanded with five tight, testable sub-clusters of new FRs:
  (1) FR-002 / FR-002a / FR-002b — pattern-based sensitive-field policy
  with CI lint for newly-discovered fields and per-attribute redaction
  unit tests; (2) FR-001 / FR-001a / FR-001b — three-source layered
  credential precedence (block > env > `credentials_file`) with
  mismatch-warning and a hard ban on importing native external-secret-store
  clients; (3) FR-003b / FR-003c — strict TLS verification with no
  skip-verify in any build, `ca_bundle` PEM as the only customization
  surface, and matching test assertions; (4) FR-032 / FR-032a / FR-032b
  — GPG + Sigstore cosign + SLSA Build L3 provenance + a
  `govulncheck`/`osv-scanner` gate that blocks releases on HIGH or
  CRITICAL CVEs; (5) FR-020a..FR-020d — provider-emitted structured
  audit log per NC2 HTTP call via `tflog` at INFO, routed via standard
  `TF_LOG` / `TF_LOG_PATH`, with mandatory per-operation test coverage.
  Checklist items are re-verified and all still pass.
