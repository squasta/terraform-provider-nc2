package dsshared

import "github.com/nutanix/terraform-provider-nc2/internal/client"

// RuntimeAccessor is the data-source-side mirror of
// orgshared.RuntimeAccessor. It is implemented by the runtime bundle
// that the provider hands to every data source via the framework's
// ProviderData field. Re-declared here (rather than imported from
// orgshared) to keep data sources free of any dependency on the
// resources tree.
type RuntimeAccessor interface {
	// NC2Client returns the configured NC2 HTTP client. May be nil
	// before the provider's Configure has run; callers should bail
	// out gracefully in that case.
	NC2Client() *client.Client
}
