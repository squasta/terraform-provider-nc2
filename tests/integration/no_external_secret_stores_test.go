package integration

import (
	"os/exec"
	"strings"
	"testing"
)

// TestNoExternalSecretStoresInBinaryDeps is the binary-graph form
// of FR-001b: the provider MUST NOT pull in a native client for any
// external secret store (Vault, AWS Secrets Manager, Azure Key
// Vault, GCP Secret Manager, 1Password) along the import graph that
// is actually linked into the provider binary.
//
// We use `go list -deps github.com/nutanix/terraform-provider-nc2`
// (NOT `... /...` and NOT `go list -m all`) so the search is scoped
// to packages reachable from `main`. Tools-only deps (kin-openapi,
// sigstore which transitively imports vault, etc.) live behind the
// `//go:build tools` constraint and are deliberately excluded.
func TestNoExternalSecretStoresInBinaryDeps(t *testing.T) {
	t.Parallel()

	cmd := exec.Command("go", "list", "-deps", "github.com/nutanix/terraform-provider-nc2")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps failed: %v\n%s", err, out)
	}
	forbidden := []string{
		"vault", "secretsmanager", "keyvault", "secret-manager", "1password",
	}
	for _, line := range strings.Split(string(out), "\n") {
		pkg := strings.ToLower(strings.TrimSpace(line))
		if pkg == "" {
			continue
		}
		for _, bad := range forbidden {
			if strings.Contains(pkg, bad) {
				t.Errorf("FR-001b violation: binary dep %q matches forbidden pattern %q", pkg, bad)
			}
		}
	}
}
