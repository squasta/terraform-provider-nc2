// Package integration hosts cross-cutting binary-level guards that
// pin invariants no individual package test can pin (FR-001b,
// FR-003c, FR-021, FR-029, SC-007).
package integration

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoInsecureSkipVerifyAnywhere is the binary-level form of
// FR-003c: assert that no Go file under the module sets
// `InsecureSkipVerify: true`. The per-package internal/client/tls_test.go
// limits its scan to that package; this test extends the same
// invariant to every other internal/* and tools/* package.
func TestNoInsecureSkipVerifyAnywhere(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)
	var found []string
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		// Test files are explicitly allowed to mention the literal
		// for documentation / search purposes (this very file does).
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(body)
		for _, bad := range []string{"InsecureSkipVerify: true", "InsecureSkipVerify:true"} {
			if strings.Contains(text, bad) {
				rel, _ := filepath.Rel(root, path)
				found = append(found, rel+" :: "+bad)
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk: %v", walkErr)
	}
	if len(found) > 0 {
		t.Errorf("FR-003c violation(s): %v", found)
	}
}
