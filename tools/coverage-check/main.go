// Command coverage-check is the CI gate for FR-021 / FR-021a.
//
// It loads `openapi/openapi.json`, walks every per-package
// `internal/{resources,datasources,actions}/*/openapi_mapping.go`
// file for `var OperationMappings = []oapi.Mapping{...}` blocks,
// and reports any operationId in the OpenAPI doc that lacks a
// covering Terraform-side mapping.
//
// Output:
//
//   - tools/coverage-check/output/coverage-report.json (per R-09 schema)
//   - On exit:
//     0 → every operation covered.
//     1 → at least one operation uncovered.
//     2 → tool error (parse failure, IO error, etc.).
//
// The tool is intentionally side-effect-light: the only writes are
// to the output JSON file; everything else is logging on stderr.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	exitOK     = 0
	exitFail   = 1
	exitErr    = 2
	defaultIn  = "openapi/openapi.json"
	defaultOut = "tools/coverage-check/output/coverage-report.json"
)

type oapiOperation struct {
	OperationID string `json:"operation_id"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	Tag         string `json:"tag,omitempty"`
}

type tfMapping struct {
	OperationID string `json:"operation_id"`
	TerraformOp string `json:"terraform_op"`
	SourceFile  string `json:"source_file"`
}

type pair struct {
	Operation oapiOperation `json:"operation"`
	MappedTo  *tfMapping    `json:"mapped_to,omitempty"`
}

type report struct {
	OpenAPISource   string          `json:"openapi_source"`
	TotalOperations int             `json:"total_operations"`
	Covered         int             `json:"covered"`
	Uncovered       []oapiOperation `json:"uncovered"`
	Extras          []tfMapping     `json:"extras"`
	Mapping         []pair          `json:"mapping"`
}

func main() {
	in := flag.String("in", defaultIn, "path to openapi.json")
	out := flag.String("out", defaultOut, "path to write coverage-report.json")
	root := flag.String("root", ".", "repo root for scanning openapi_mapping.go files")
	flag.Parse()

	rep, code, err := run(*in, *out, *root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "coverage-check: error:", err)
		os.Exit(exitErr)
	}
	if code == exitOK {
		fmt.Fprintf(os.Stderr, "coverage-check: %d/%d operations covered\n", rep.Covered, rep.TotalOperations)
	} else {
		fmt.Fprintf(os.Stderr, "coverage-check: FAIL %d uncovered (extras=%d)\n", len(rep.Uncovered), len(rep.Extras))
		for _, u := range rep.Uncovered {
			fmt.Fprintf(os.Stderr, "  uncovered: %s %s [%s]\n", u.Method, u.Path, u.OperationID)
		}
	}
	os.Exit(code)
}

// run is the testable workhorse. Returns the report, an exit code,
// and a tool-level error (distinct from "uncovered operations
// exist", which is signalled via exit code 1).
func run(in, out, root string) (*report, int, error) {
	ops, err := loadOperations(in)
	if err != nil {
		return nil, exitErr, fmt.Errorf("load operations: %w", err)
	}
	mappings, err := scanMappings(root)
	if err != nil {
		return nil, exitErr, fmt.Errorf("scan mappings: %w", err)
	}
	rep := buildReport(in, ops, mappings)

	if err := writeReport(out, rep); err != nil {
		return nil, exitErr, fmt.Errorf("write report: %w", err)
	}

	if len(rep.Uncovered) > 0 {
		return rep, exitFail, nil
	}
	return rep, exitOK, nil
}

// loadOperations parses the OpenAPI doc and returns one
// oapiOperation per (path, method).
//
// Implementation note: we use stdlib `encoding/json` rather than a
// strict OpenAPI parser because the upstream NC2 spec contains a
// handful of non-canonical `items: "<RefName>"` strings (which kin-
// openapi rightly rejects). Since we only need `paths/<route>/<method>/operationId`,
// stdlib JSON gets us there without forcing spec compliance on a
// vendored input we don't control.
func loadOperations(path string) ([]oapiOperation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc struct {
		Paths map[string]map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	httpMethods := map[string]bool{
		"get": true, "post": true, "put": true, "patch": true, "delete": true,
		"options": true, "head": true, "trace": true,
	}
	var ops []oapiOperation
	for routePath, pathItem := range doc.Paths {
		for method, raw := range pathItem {
			lc := strings.ToLower(method)
			if !httpMethods[lc] {
				continue
			}
			op, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			id, _ := op["operationId"].(string)
			tag := ""
			if tags, ok := op["tags"].([]any); ok && len(tags) > 0 {
				if first, ok := tags[0].(string); ok {
					tag = first
				}
			}
			ops = append(ops, oapiOperation{
				OperationID: id,
				Method:      strings.ToUpper(lc),
				Path:        routePath,
				Tag:         tag,
			})
		}
	}
	sort.Slice(ops, func(i, j int) bool {
		if ops[i].Path != ops[j].Path {
			return ops[i].Path < ops[j].Path
		}
		return ops[i].Method < ops[j].Method
	})
	return ops, nil
}

// scanMappings walks the repo for `openapi_mapping.go` files and
// extracts every `OperationMappings` literal element pair.
func scanMappings(root string) ([]tfMapping, error) {
	var mappings []tfMapping
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			// Skip vendored / build / non-source dirs.
			name := d.Name()
			if name == "vendor" || name == ".git" || name == "tools" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "openapi_mapping.go" {
			return nil
		}
		ms, err := parseOpenAPIMappingFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		// Make path repo-relative for nicer report output.
		rel, _ := filepath.Rel(root, path)
		for i := range ms {
			ms[i].SourceFile = rel
		}
		mappings = append(mappings, ms...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mappings, nil
}

// parseOpenAPIMappingFile parses a single openapi_mapping.go file,
// finds the OperationMappings = []oapi.Mapping{...} declaration,
// and extracts each {OperationID: ..., TerraformOp: ...} pair.
func parseOpenAPIMappingFile(path string) ([]tfMapping, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	var out []tfMapping
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vspec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vspec.Names {
				if name.Name != "OperationMappings" {
					continue
				}
				if i >= len(vspec.Values) {
					continue
				}
				lit, ok := vspec.Values[i].(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, elt := range lit.Elts {
					m, ok := extractMapping(elt)
					if ok {
						out = append(out, m)
					}
				}
			}
		}
	}
	return out, nil
}

func extractMapping(node ast.Expr) (tfMapping, bool) {
	cl, ok := node.(*ast.CompositeLit)
	if !ok {
		return tfMapping{}, false
	}
	var m tfMapping
	for _, elt := range cl.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		val, ok := kv.Value.(*ast.BasicLit)
		if !ok || val.Kind != token.STRING {
			continue
		}
		// Strip surrounding quotes.
		v := strings.Trim(val.Value, "\"`")
		switch key.Name {
		case "OperationID":
			m.OperationID = v
		case "TerraformOp":
			m.TerraformOp = v
		}
	}
	if m.OperationID == "" || m.TerraformOp == "" {
		return tfMapping{}, false
	}
	return m, true
}

// buildReport joins the OpenAPI inventory with the discovered
// mappings, classifying each operation as covered / uncovered and
// each mapping as either matched or extra (mapping that points to
// an operationId not present in the spec).
func buildReport(source string, ops []oapiOperation, mappings []tfMapping) *report {
	covers := make(map[string][]tfMapping, len(mappings))
	for _, m := range mappings {
		covers[m.OperationID] = append(covers[m.OperationID], m)
	}

	rep := &report{
		OpenAPISource:   source,
		TotalOperations: len(ops),
	}
	seenIDs := make(map[string]struct{})
	for _, op := range ops {
		seenIDs[op.OperationID] = struct{}{}
		matches, ok := covers[op.OperationID]
		if !ok || len(matches) == 0 {
			rep.Uncovered = append(rep.Uncovered, op)
			rep.Mapping = append(rep.Mapping, pair{Operation: op})
			continue
		}
		rep.Covered++
		// Pin the first mapping (sorted ascending by source file +
		// terraform op) for deterministic output.
		sort.Slice(matches, func(i, j int) bool {
			if matches[i].SourceFile != matches[j].SourceFile {
				return matches[i].SourceFile < matches[j].SourceFile
			}
			return matches[i].TerraformOp < matches[j].TerraformOp
		})
		m := matches[0]
		rep.Mapping = append(rep.Mapping, pair{Operation: op, MappedTo: &m})
	}
	for _, m := range mappings {
		if _, ok := seenIDs[m.OperationID]; !ok {
			rep.Extras = append(rep.Extras, m)
		}
	}
	sort.Slice(rep.Extras, func(i, j int) bool {
		return rep.Extras[i].OperationID < rep.Extras[j].OperationID
	})
	return rep
}

func writeReport(path string, rep *report) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	body, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0o644)
}
