package aws_cluster //nolint:revive,staticcheck

// Sensitive returns the per-resource sensitive registry entries
// required by FR-002 for `nc2_aws_cluster`. Per data-model.md §4
// the IP-address lists inside the access policies are flagged as
// operational secrets even though they don't match the FR-002 name
// patterns.
func Sensitive() []string {
	return []string{
		"network.management_services_access_policy.ip_addresses",
		"network.prism_element_access_policy.ip_addresses",
		"access_policy.ip_addresses",
	}
}
