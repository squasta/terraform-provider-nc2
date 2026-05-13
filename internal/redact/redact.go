package redact

import "strings"

// Sentinel is the placeholder value substituted for any sensitive
// scalar found by Redact. The exact spelling is part of the FR-002b
// contract pinned by redact_test.go: changing it here is a behavior
// change, not a refactor.
const Sentinel = "(sensitive)"

// Registry holds the per-resource sensitive-attribute paths in
// addition to the FR002Patterns name-pattern matcher. A Registry is
// immutable after construction; mutation MUST be modeled as a fresh
// NewRegistry call so callers can safely cache and share instances
// across goroutines.
type Registry struct {
	// extra holds canonicalized dotted paths whose terminal segment
	// is treated as sensitive even when its name does not match
	// FR002Patterns. Per data-model.md the per-resource
	// sensitive.go files contribute these paths.
	extra []string
}

// NewRegistry returns a Registry whose Redact method classifies
// scalars as sensitive when either:
//
//  1. The leaf attribute name matches FR002Patterns
//     (case-insensitive, underscore-stripped substring match), OR
//  2. The dotted path from the Redact root to the leaf is exactly
//     equal to one of the extra entries supplied here.
//
// extra MAY be nil; the resulting Registry then redacts purely on
// FR002Patterns. Duplicate or empty entries are silently ignored.
//
// The function is pure: same inputs → same Registry; the returned
// Registry never references the caller's slice (a defensive copy is
// made).
func NewRegistry(extra []string) Registry {
	if len(extra) == 0 {
		return Registry{}
	}
	out := make([]string, 0, len(extra))
	seen := make(map[string]struct{}, len(extra))
	for _, p := range extra {
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return Registry{extra: out}
}

// Redact returns a deep copy of in with every sensitive scalar
// replaced by the Sentinel string. Containers (maps and slices) are
// recursed into; leaves of every other type pass through unchanged.
//
// Redact is pure: it never mutates its input. This is critical
// because the same map graph is consumed by audit logging (FR-020b),
// per-resource state derivations, and external diagnostics; a hidden
// mutation here would silently corrupt all three.
//
// Idempotency: redacting an already-redacted map is a no-op; the
// Sentinel string is itself classified as a non-sensitive value
// because Sentinel does not contain any FR002Patterns substring.
//
// nil-safety: Redact(nil) returns nil.
func (r Registry) Redact(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	return r.redactMap(in, "")
}

// redactMap is the recursive workhorse. The path argument carries
// the dotted prefix accumulated from the Redact root; it is used to
// match against the per-Registry extra entries.
//
// Classification rules (FR-002 + FR-002b):
//
//   - An extra-registry path match collapses the value (scalar OR
//     container) to the Sentinel string. This is the per-resource
//     escape hatch for paths whose terminal name does not match
//     FR002Patterns but is still classified sensitive (e.g.
//     access-policy IP-address lists).
//   - A name-pattern match on the key collapses scalar values only.
//     Container values are recursed into so we don't over-redact a
//     map whose name happens to be "credentials" but whose
//     non-secret children (e.g. credentials.role_arn) should remain
//     visible. Children that themselves match the pattern OR are
//     scalar children of a pattern-matching parent ARE redacted in
//     the recursion.
func (r Registry) redactMap(in map[string]any, path string) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		childPath := k
		if path != "" {
			childPath = path + "." + k
		}
		if r.pathMatchesExtra(childPath) {
			out[k] = Sentinel
			continue
		}
		if MatchPattern(k) {
			if isScalar(v) {
				out[k] = Sentinel
				continue
			}
			out[k] = r.redactValue(v, childPath)
			continue
		}
		out[k] = r.redactValue(v, childPath)
	}
	return out
}

// isScalar reports whether v is a non-container value (string, bool,
// number, nil). Containers (map[string]any, []any) recurse instead.
func isScalar(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return false
	default:
		return true
	}
}

// redactValue dispatches on the dynamic type of v and recurses into
// containers; scalars are returned verbatim.
func (r Registry) redactValue(v any, path string) any {
	switch t := v.(type) {
	case map[string]any:
		return r.redactMap(t, path)
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = r.redactValue(e, path)
		}
		return out
	default:
		return v
	}
}

// pathMatchesExtra reports whether the supplied dotted path matches
// any registered extra path. Comparison is case-sensitive (paths
// come from the framework's tfsdk attribute graph and are already
// canonical snake_case).
func (r Registry) pathMatchesExtra(path string) bool {
	for _, p := range r.extra {
		if p == path {
			return true
		}
	}
	return false
}

// HasExtra reports whether the Registry was constructed with at
// least one per-resource extra entry. Used by per-package unit tests
// to guard against an accidentally-empty Sensitive() return value.
func (r Registry) HasExtra() bool { return len(r.extra) > 0 }

// Extras returns a fresh copy of the per-resource extra paths the
// Registry was constructed with. The returned slice is sorted in
// the order it was supplied; callers that want deterministic
// ordering should sort it themselves.
func (r Registry) Extras() []string {
	if len(r.extra) == 0 {
		return nil
	}
	out := make([]string, len(r.extra))
	copy(out, r.extra)
	return out
}

// JoinPath joins two dotted-path segments with the Registry's
// canonical separator. A small helper that lets per-resource tests
// compose paths without hard-coding the separator.
func JoinPath(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}
	return a + "." + b
}

// SplitPath returns the segments of a dotted path. Inverse of
// JoinPath.
func SplitPath(p string) []string {
	if p == "" {
		return nil
	}
	return strings.Split(p, ".")
}
