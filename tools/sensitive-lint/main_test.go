package main

import (
	"os"
	"path/filepath"
	"testing"
)

const synthOpenAPI = `{
  "openapi": "3.0.0",
  "info": {"title": "synth", "version": "1"},
  "paths": {},
  "components": {
    "schemas": {
      "Cred": {
        "type": "object",
        "properties": {
          "api_key":   {"type": "string"},
          "name":      {"type": "string"},
          "secret":    {"type": "string"},
          "role_arn":  {"type": "string"}
        }
      }
    }
  }
}`

// TestRun_NonStrictReportsButExits0 makes sure FR-002a's
// informational mode works.
func TestRun_NonStrictReportsButExits0(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	oapi := filepath.Join(root, "openapi.json")
	if err := os.WriteFile(oapi, []byte(synthOpenAPI), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	out := filepath.Join(root, "report.json")
	rep, code, err := run(oapi, out, root, false)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if code != exitOK {
		t.Errorf("non-strict should exit 0; got %d", code)
	}
	if rep.PatternHits != 2 {
		// api_key + secret. "name" and "role_arn" are not patterned.
		t.Errorf("PatternHits=%d; want 2 (api_key, secret)", rep.PatternHits)
	}
	if len(rep.Findings) == 0 {
		t.Errorf("expected at least one MEDIUM finding describing unclassified pattern hits")
	}
}

// TestRun_StrictFailsOnFindings.
func TestRun_StrictFailsOnFindings(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	oapi := filepath.Join(root, "openapi.json")
	if err := os.WriteFile(oapi, []byte(synthOpenAPI), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, code, err := run(oapi, filepath.Join(root, "rep.json"), root, true)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if code != exitFail {
		t.Errorf("strict mode should fail on findings; got code %d", code)
	}
}

// TestRun_RegistryEntryClassifiesAttribute clears one MEDIUM
// finding by registering api_key in a fake sensitive.go.
func TestRun_RegistryEntryClassifiesAttribute(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	oapi := filepath.Join(root, "openapi.json")
	if err := os.WriteFile(oapi, []byte(synthOpenAPI), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	dir := filepath.Join(root, "internal", "resources", "synth")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	body := `package synth

func Sensitive() []string {
    return []string{
        "credentials.api_key",
        "credentials.secret",
    }
}
`
	if err := os.WriteFile(filepath.Join(dir, "sensitive.go"), []byte(body), 0o644); err != nil {
		t.Fatalf("write sensitive.go: %v", err)
	}
	rep, code, err := run(oapi, filepath.Join(root, "rep.json"), root, false)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if code != exitOK {
		t.Errorf("expected non-strict OK; got %d", code)
	}
	if rep.Classified != 2 {
		t.Errorf("expected both pattern hits classified via registry; Classified=%d", rep.Classified)
	}
	if len(rep.Findings) != 0 {
		t.Errorf("expected no findings when both registry-covered; got %+v", rep.Findings)
	}
}

// TestParseSensitiveFile extracts the string literals.
func TestParseSensitiveFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "sensitive.go")
	body := `package synth

func Sensitive() []string {
    return []string{
        "a.b.c",
        "x",
    }
}

func unrelated() string { return "ignore.me" }
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	out, err := parseSensitiveFile(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(out) != 2 || out[0] != "a.b.c" || out[1] != "x" {
		t.Errorf("got %+v", out)
	}
}
