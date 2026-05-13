// Package redact implements the FR-002 sensitive-field policy:
//
//   - A maintained pattern list (FR002Patterns) that, by case-insensitive
//     substring match, classifies attribute names as sensitive. Any attribute
//     whose name contains one of these substrings is treated as sensitive in
//     plan output, apply output, persisted state, and debug logs (FR-002b).
//
//   - A per-resource sensitive registry (Registry) that the
//     internal/resources/<type>/sensitive.go files contribute to, capturing
//     sensitive attribute paths whose names do NOT match the pattern list.
//
// The package is intentionally pure: no I/O, no globals beyond the FR-002
// pattern list itself, no dependency on the framework or on logging. This
// makes redaction deterministic and trivially testable, and lets the package
// be reused by:
//   - internal/audit (FR-020b — redact every audit-record field).
//   - tools/sensitive-lint (FR-002a — same pattern semantics on the CLI).
//   - per-resource unit tests (FR-002b — assert no plaintext leaks).
//
// See specs/001-nutanix-nc2-provider/spec.md "FR-002" for the authoritative
// policy and specs/001-nutanix-nc2-provider/data-model.md for which
// attributes are registered per resource.
package redact

import "strings"

// FR002Patterns is the case-insensitive substring list defined by FR-002.
//
// The list is intentionally short and stable: changes here are an FR-002
// amendment, not a routine refactor. Adding or removing an entry MUST be
// accompanied by:
//
//  1. An update to the spec (FR-002).
//  2. A documented `[Spec Kit] update FR-002 patterns` commit.
//  3. A re-run of `make sensitive-lint-strict` in case the change re-classifies
//     existing OpenAPI attributes (FR-002a).
//
// Order is not semantically significant; tests treat the list as a set.
var FR002Patterns = []string{
	"credential",
	"password",
	"secret",
	"token",
	"private_key",
	"api_key",
}

// MatchPattern reports whether name contains any FR002Patterns entry as a
// case-insensitive substring.
//
// Before comparison the input and each pattern are normalized: lowercased and
// stripped of underscores. Normalization is what lets a single pattern entry
// catch every realistic spelling of the same concept across the codebase:
//
//   - NC2 JSON field names use snake_case ("api_key", "private_key").
//   - Go struct field names use PascalCase ("ApiKey", "PrivateKey", "APIKey").
//   - Terraform schema attribute paths use snake_case again ("api_key").
//
// Treating "_" as semantically empty unifies these spellings. The
// underscore-stripping rule is part of the FR-002 contract pinned by
// patterns_test.go::TestMatchPattern_PositiveCases.
//
// The function is pure: same input → same output, no side effects. It is the
// hot-path classifier used by the audit emitter and by the per-resource
// sensitive registries; correctness here is critical because the consequence
// of a false negative is a credential leak.
//
// The empty string is never sensitive.
//
// Examples:
//
//	MatchPattern("api_key")        == true
//	MatchPattern("APIKey")         == true   // case-insensitive
//	MatchPattern("credentialed")   == true   // substring, per FR-002
//	MatchPattern("description")    == false
//	MatchPattern("")               == false
func MatchPattern(name string) bool {
	if name == "" {
		return false
	}
	norm := normalize(name)
	for _, p := range FR002Patterns {
		if strings.Contains(norm, normalize(p)) {
			return true
		}
	}
	return false
}

// normalize lowercases s and removes underscores. See MatchPattern for the
// rationale. Pure function.
func normalize(s string) string {
	lower := strings.ToLower(s)
	if !strings.ContainsRune(lower, '_') {
		return lower
	}
	return strings.ReplaceAll(lower, "_", "")
}
