package organization

// Sensitive returns the per-resource list of dotted attribute paths
// that the redactor must replace with `(sensitive)` even when the
// attribute name does not match a built-in FR-002 pattern.
//
// nc2_organization has no sensitive fields per data-model.md §1.
// The function is kept exported so tools/sensitive-lint can read it
// uniformly across every resource package without special-casing
// "no sensitive fields".
func Sensitive() []string { return []string{} }
