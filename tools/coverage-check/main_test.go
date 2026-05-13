package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const synthOpenAPI = `{
  "openapi": "3.0.0",
  "info": {"title": "synth", "version": "1"},
  "paths": {
    "/foo": {
      "get":  {"operationId": "synth.Foo.list",   "tags": ["foo"], "responses": {"200": {"description": "ok"}}},
      "post": {"operationId": "synth.Foo.create", "tags": ["foo"], "responses": {"201": {"description": "ok"}}}
    },
    "/bar/{id}": {
      "get":  {"operationId": "synth.Bar.show",   "tags": ["bar"], "responses": {"200": {"description": "ok"}}}
    }
  }
}`

const fullyCoveredMapping = `package x

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

var OperationMappings = []oapi.Mapping{
    {OperationID: "synth.Foo.list",   TerraformOp: "data.synth_foo.Read"},
    {OperationID: "synth.Foo.create", TerraformOp: "synth_foo.Create"},
    {OperationID: "synth.Bar.show",   TerraformOp: "synth_bar.Read"},
}
`

const partialMapping = `package x

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

var OperationMappings = []oapi.Mapping{
    {OperationID: "synth.Foo.list",   TerraformOp: "data.synth_foo.Read"},
}
`

const extraMapping = `package x

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

var OperationMappings = []oapi.Mapping{
    {OperationID: "synth.Foo.list",       TerraformOp: "data.synth_foo.Read"},
    {OperationID: "synth.Foo.create",     TerraformOp: "synth_foo.Create"},
    {OperationID: "synth.Bar.show",       TerraformOp: "synth_bar.Read"},
    {OperationID: "synth.Phantom.unknown", TerraformOp: "synth_phantom.Read"},
}
`

// writeFixture builds a temporary repo skeleton with the supplied
// OpenAPI doc and openapi_mapping.go file under
// internal/resources/synth/ so the scanner finds it.
func writeFixture(t *testing.T, openapiBody, mappingBody string) (root, oapi, out string) {
	t.Helper()
	root = t.TempDir()
	oapi = filepath.Join(root, "openapi.json")
	if err := os.WriteFile(oapi, []byte(openapiBody), 0o644); err != nil {
		t.Fatalf("write openapi: %v", err)
	}
	dir := filepath.Join(root, "internal", "resources", "synth")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "openapi_mapping.go"), []byte(mappingBody), 0o644); err != nil {
		t.Fatalf("write mapping: %v", err)
	}
	out = filepath.Join(root, "report.json")
	return
}

// TestRun_FullyCovered exits 0.
func TestRun_FullyCovered(t *testing.T) {
	t.Parallel()

	root, oapi, out := writeFixture(t, synthOpenAPI, fullyCoveredMapping)
	rep, code, err := run(oapi, out, root)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if code != exitOK {
		t.Errorf("exit code = %d; want %d", code, exitOK)
	}
	if rep.Covered != 3 || rep.TotalOperations != 3 {
		t.Errorf("rep = %+v", rep)
	}
	if len(rep.Uncovered) != 0 || len(rep.Extras) != 0 {
		t.Errorf("expected no uncovered or extras; got %+v", rep)
	}
	// File must exist and parse as the right schema.
	body, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var parsed report
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Errorf("report not valid JSON: %v", err)
	}
}

// TestRun_PartialCoverageFails returns exit 1 + reports uncovered.
func TestRun_PartialCoverageFails(t *testing.T) {
	t.Parallel()

	root, oapi, out := writeFixture(t, synthOpenAPI, partialMapping)
	rep, code, err := run(oapi, out, root)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if code != exitFail {
		t.Errorf("exit code = %d; want %d", code, exitFail)
	}
	if len(rep.Uncovered) != 2 {
		t.Errorf("expected 2 uncovered; got %d (%+v)", len(rep.Uncovered), rep.Uncovered)
	}
}

// TestRun_DetectsExtras flags mappings with no spec counterpart.
func TestRun_DetectsExtras(t *testing.T) {
	t.Parallel()

	root, oapi, out := writeFixture(t, synthOpenAPI, extraMapping)
	rep, _, err := run(oapi, out, root)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(rep.Extras) != 1 || rep.Extras[0].OperationID != "synth.Phantom.unknown" {
		t.Errorf("extras = %+v", rep.Extras)
	}
}

// TestParseOpenAPIMappingFile_IgnoresUnrelatedDecls makes sure the
// parser only picks up the canonical OperationMappings name.
func TestParseOpenAPIMappingFile_IgnoresUnrelatedDecls(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "openapi_mapping.go")
	body := `package x

import "github.com/nutanix/terraform-provider-nc2/internal/oapi"

var SomethingElse = []oapi.Mapping{
    {OperationID: "should.not.match", TerraformOp: "x.Read"},
}

var OperationMappings = []oapi.Mapping{
    {OperationID: "real.op", TerraformOp: "real.Read"},
}
`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	out, err := parseOpenAPIMappingFile(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(out) != 1 || out[0].OperationID != "real.op" {
		t.Errorf("got %+v", out)
	}
}
