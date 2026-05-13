// Package orgshared houses the small set of cross-package types that
// the per-resource and per-action packages need to extract the
// configured NC2 client from the framework's `ProviderData` bundle
// during Configure.
//
// Keeping this in its own package — rather than importing
// `internal/provider` — avoids an import cycle: `internal/provider`
// itself wires up the per-resource constructors and so cannot be
// depended on by them.
package orgshared

import "github.com/nutanix/terraform-provider-nc2/internal/client"

// RuntimeAccessor is implemented by the runtime bundle that the
// provider hands to every resource / data source / action via the
// framework's ProviderData field. The single getter returns the
// configured NC2 HTTP client; ImportState/Read/etc. then call methods
// on it.
//
// The interface lives here, in a leaf package, so that every consumer
// can `import orgshared` without pulling in the provider implementation.
type RuntimeAccessor interface {
	// NC2Client returns the configured NC2 HTTP client. May be nil
	// if the provider has not been configured (e.g. during early
	// validation), in which case callers should add a clear
	// diagnostic and bail out of the operation.
	NC2Client() *client.Client
}
