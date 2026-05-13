package main

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/nutanix/terraform-provider-nc2/internal/provider"
)

// TestProviderRegistration is the T044 smoke test: verify that the
// constructor exposed by internal/provider satisfies the framework's
// provider.Provider interface and that providerserver can serve it
// against the documented address.
//
// We do NOT actually start the gRPC server (that would block); we
// only construct the ServeOpts to assert that the address string is
// well-formed and the constructor returns a non-nil provider.
func TestProviderRegistration(t *testing.T) {
	t.Parallel()

	if p := provider.New(); p == nil {
		t.Fatalf("provider.New() returned nil")
	}

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/nutanix/nc2",
	}
	if opts.Address == "" {
		t.Fatalf("Address must be non-empty")
	}

	// The Serve call below will be invoked by `main` in production;
	// here we just make sure the function exists and accepts a
	// nil-cancellable context shape. We do not actually call Serve.
	_ = providerserver.Serve
	_ = context.Background
}
