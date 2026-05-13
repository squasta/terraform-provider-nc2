package provider

import (
	"os/exec"
	"strings"
	"testing"
)

// forbiddenSecretStorePatterns lists substrings whose presence in the
// provider's transitive module dependency graph violates FR-001b
// ("no native external secret-store imports"). Match is
// case-insensitive substring on the module path.
var forbiddenSecretStorePatterns = []string{
	"vault",
	"secretsmanager",
	"keyvault",
	"secret-manager",
	"1password",
}

// TestNoExternalSecretStoreImports walks the entire `go list -deps`
// graph rooted at the main package and fails if any module path
// matches one of forbiddenSecretStorePatterns.
//
// Per FR-001b, the provider MUST NOT import a native client for any
// external secret store; credentials arrive already-resolved from
// internal/provider.Resolve.
//
// The test is deliberately implemented as `go list` (rather than a
// hand-rolled module tree walk) so we get the same view the linker
// has — including conditional imports added by build tags.
func TestNoExternalSecretStoreImports(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("go", "list", "-deps", "github.com/nutanix/terraform-provider-nc2/...")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps failed: %v\n%s", err, out)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" {
			continue
		}
		for _, bad := range forbiddenSecretStorePatterns {
			if strings.Contains(line, bad) {
				t.Errorf("FR-001b violation: dependency %q matches forbidden pattern %q", line, bad)
			}
		}
	}
}
