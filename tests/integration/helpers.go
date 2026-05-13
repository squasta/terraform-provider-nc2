package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot returns the absolute path to the repo root by asking
// `git rev-parse --show-toplevel`. Cached per test process.
//
// We cannot use t.TempDir based fixtures here; integration tests
// run against the real repo on disk.
func repoRoot(t *testing.T) string {
	t.Helper()
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		// Fallback: current dir + walk upward to module root.
		wd, _ := os.Getwd()
		for cur := wd; cur != "/"; cur = filepath.Dir(cur) {
			if fileExistsHelper(filepath.Join(cur, "go.mod")) {
				return cur
			}
		}
		t.Fatalf("repoRoot: %v", err)
	}
	return strings.TrimSpace(string(out))
}

// fileExistsHelper duplicates fileExists to avoid Go-level symbol
// shadowing across files in this package.
func fileExistsHelper(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// readFile is a tiny test-only wrapper around os.ReadFile that
// accepts a path and returns the bytes, suitable for chained use in
// table-driven tests.
func readFile(p string) ([]byte, error) {
	return os.ReadFile(p)
}
