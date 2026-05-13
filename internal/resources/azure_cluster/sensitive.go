package azure_cluster //nolint:revive,staticcheck

// Sensitive returns the FR-002 per-resource registry for
// `nc2_azure_cluster`. Mirrors AWS minus the AWS-only access policy.
func Sensitive() []string {
	return []string{
		"network.management_services_access_policy.ip_addresses",
		"network.prism_element_access_policy.ip_addresses",
	}
}
