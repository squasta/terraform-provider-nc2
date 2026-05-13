package integration

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestOpenAPICoverage_Real is the SC-001 gate: tools/coverage-check
// MUST exit 0 against the committed openapi/openapi.json and the
// emitted report MUST have empty `uncovered`.
//
// This is the integration-level version of
// tools/coverage-check/main_test.go (which uses synthetic fixtures).
func TestOpenAPICoverage_Real(t *testing.T) {
	root := repoRoot(t)
	out := filepath.Join(t.TempDir(), "coverage-report.json")

	cmd := exec.Command("go", "run", "./tools/coverage-check",
		"-in", filepath.Join(root, "openapi/openapi.json"),
		"-out", out,
		"-root", root,
	)
	cmd.Dir = root
	combined, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("coverage-check exit code != 0: %v\n%s", err, combined)
	}

	body, err := readFile(out)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var rep struct {
		TotalOperations int   `json:"total_operations"`
		Covered         int   `json:"covered"`
		Uncovered       []any `json:"uncovered"`
	}
	if err := json.Unmarshal(body, &rep); err != nil {
		t.Fatalf("parse report: %v\n%s", err, body)
	}
	if len(rep.Uncovered) > 0 {
		t.Errorf("expected zero uncovered operations; got %d", len(rep.Uncovered))
	}
	if rep.Covered != rep.TotalOperations {
		t.Errorf("Covered=%d != TotalOperations=%d", rep.Covered, rep.TotalOperations)
	}
}
