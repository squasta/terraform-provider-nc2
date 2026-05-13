package gcp_cluster //nolint:revive,staticcheck

// Sensitive returns the FR-002 per-resource registry for
// `nc2_gcp_cluster`.
func Sensitive() []string {
	return []string{
		"network.management_services_access_policy.ip_addresses",
		"network.prism_element_access_policy.ip_addresses",
	}
}
