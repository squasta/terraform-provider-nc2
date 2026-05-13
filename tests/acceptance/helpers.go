// Package acceptance houses the FR-023..FR-026 acceptance test
// suite. Every test in this package is gated by `TF_ACC=1`; absent
// the gate, tests skip rather than fail, so unit-test CI lanes can
// run them without provisioning real cloud credentials.
package acceptance

import (
	"context"
	"math/rand/v2"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/nutanix/terraform-provider-nc2/internal/provider"
)

// envAccGate is the env var that opts a CI lane in to acceptance
// tests. Without it set to "1", every test in this package skips.
const envAccGate = "TF_ACC"

// SkipIfNotAcc skips t when TF_ACC is not set to "1". Mirrors the
// semantics of terraform-plugin-testing's gate.
func SkipIfNotAcc(t *testing.T) {
	t.Helper()
	if os.Getenv(envAccGate) != "1" {
		t.Skip("acceptance test skipped: set TF_ACC=1 to run")
	}
}

// SkipIfMissing skips t when any of the named environment variables
// are unset or empty. Returns the resolved values in the same order
// they were supplied so callers can use a positional slice.
func SkipIfMissing(t *testing.T, names ...string) []string {
	t.Helper()
	values := make([]string, len(names))
	var missing []string
	for i, n := range names {
		v := os.Getenv(n)
		if v == "" {
			missing = append(missing, n)
		}
		values[i] = v
	}
	if len(missing) > 0 {
		t.Skipf("acceptance test skipped: missing env vars %s", strings.Join(missing, ", "))
	}
	return values
}

// randInt returns a positive random integer suitable for embedding
// in unique test resource names. Uses math/rand/v2 (no seeding
// required; per-process RNG).
func randInt() int { return int(rand.Uint32() & 0x7fffffff) }

// ProtoV6ProviderFactories wires the in-process provider into the
// terraform-plugin-testing harness. Every test case sets this on
// resource.TestCase.
//
//nolint:gochecknoglobals // shared per-package factory map.
var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"nc2": providerserver.NewProtocol6WithError(provider.New()),
}

var _ = context.Background // imported for transitive use in subtests
