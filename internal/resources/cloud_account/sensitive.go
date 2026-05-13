package cloud_account //nolint:revive,staticcheck

// Sensitive returns the per-resource list of dotted paths the
// redactor must replace with `(sensitive)` even when the attribute
// name does not match a built-in FR-002 pattern.
//
// The full `credentials` subtree is auto-classified by the FR-002
// `credential` substring pattern. The explicit per-cloud entries
// below are belt-and-braces for spec FR-002 traceability and for
// tools/sensitive-lint to surface drift if the OpenAPI shape
// changes.
func Sensitive() []string {
	return []string{
		"credentials.aws.access_key_id",
		"credentials.aws.secret_access_key",
		"credentials.azure.client_secret",
		"credentials.gcp.service_account_json",
	}
}
