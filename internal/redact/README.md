# `internal/redact`

Pure-function library implementing the FR-002 sensitive-field policy.

## Purpose

Single source of truth for "is this attribute sensitive?" decisions. Used by:

- `internal/audit` — to redact audit-record fields before they hit `tflog`.
- `tools/sensitive-lint` — to enforce per-resource sensitive registries against the OpenAPI surface.
- per-resource unit tests — to assert no plaintext leaks (FR-002b).

## Public API

| Symbol | Kind | Purpose |
|---|---|---|
| `FR002Patterns` | `var []string` | The 6-entry case-insensitive substring list mandated by FR-002. |
| `MatchPattern(name string) bool` | func | Reports whether `name` matches FR-002 by underscore-stripped, lowercased substring match. |
| `Sentinel` | `const string` | The placeholder value substituted for redacted scalars: `"(sensitive)"`. |
| `Registry` | struct | Immutable per-package sensitive-field classifier. |
| `NewRegistry(extra []string) Registry` | func | Constructor; extra is the per-resource registry of dotted paths. |
| `(Registry).Redact(in map[string]any) map[string]any` | method | Returns a deep copy of `in` with sensitive scalars replaced by `Sentinel`. |
| `(Registry).Extras() []string` | method | Returns a fresh copy of the per-resource extra paths. |
| `(Registry).HasExtra() bool` | method | Reports whether the registry has at least one extra entry. |
| `JoinPath(a, b string) string` | func | Helper to compose dotted paths. |
| `SplitPath(p string) []string` | func | Inverse of `JoinPath`. |

## FR-002 pattern list

Case-insensitive substring match (after lowercasing and stripping underscores):

- `credential`
- `password`
- `secret`
- `token`
- `private_key`
- `api_key`

Adding or removing a pattern is an FR-002 amendment; it MUST be paired with:

1. A spec update.
2. A re-run of `make sensitive-lint-strict`.

## Classification rules

For each scalar value found while walking the input map:

1. If the dotted path from the `Redact` root to the leaf matches a per-resource
   registry entry exactly → replace value with `Sentinel`.
2. Else if the leaf's key matches `FR002Patterns` → replace scalar with
   `Sentinel`. Container values (maps, slices) are recursed into rather than
   collapsed, to avoid over-redacting non-secret children inside a
   pattern-matching parent (e.g. `credentials.role_arn`).
3. Else → leave unchanged.

`Redact` is pure: it never mutates input. Idempotent: redacting an already-
redacted map produces an equal output.
