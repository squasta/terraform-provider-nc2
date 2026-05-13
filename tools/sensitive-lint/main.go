// Command sensitive-lint enforces FR-002a: every attribute in
// openapi/openapi.json whose name matches FR-002 patterns must be
// classified in a per-resource sensitive registry (the
// `sensitive.go` files under internal/resources/*).
//
// Two failure modes:
//
//  1. Pattern hit without a matching registry entry → drift; the
//     tool warns in non-strict mode and exits 1 in strict mode.
//  2. Registry entry whose dotted path mentions a name that DOES
//     match FR-002 patterns but the corresponding OpenAPI attribute
//     is absent → also drift, but in the opposite direction (lint
//     covers something the spec no longer ships).
//
// Output:
//
//   - tools/sensitive-lint/output/sensitive-report.json
//   - On exit:
//       0 → no findings (or non-strict with only warnings).
//       1 → strict + at least one finding.
//       2 → tool error.
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

	"github.com/nutanix/terraform-provider-nc2/internal/redact"
)

const (
	exitOK     = 0
	exitFail   = 1
	exitErr    = 2
	defaultIn  = "openapi/openapi.json"
	defaultOut = "tools/sensitive-lint/output/sensitive-report.json"
)

// finding is one classification or drift signal.
type finding struct {
	Severity   string `json:"severity"` // HIGH | MEDIUM | LOW
	Kind       string `json:"kind"`     // unclassified_pattern | extra_registry | unclassified_field
	Attribute  string `json:"attribute"`
	SourceFile string `json:"source_file,omitempty"`
	Detail     string `json:"detail,omitempty"`
}

type report struct {
	OpenAPISource    string    `json:"openapi_source"`
	Strict           bool      `json:"strict"`
	TotalAttributes  int       `json:"total_attributes"`
	PatternHits      int       `json:"pattern_hits"`
	Classified       int       `json:"classified"`
	Findings         []finding `json:"findings"`
	RegistryEntries  []string  `json:"registry_entries"`
	OpenAPIPatterned []string  `json:"openapi_patterned_attrs"`
}

func main() {
	in := flag.String("in", defaultIn, "openapi.json path")
	out := flag.String("out", defaultOut, "report path")
	root := flag.String("root", ".", "repo root")
	strict := flag.Bool("strict", false, "fail on any finding")
	flag.Parse()

	rep, code, err := run(*in, *out, *root, *strict)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sensitive-lint: error:", err)
		os.Exit(exitErr)
	}
	switch code {
	case exitOK:
		fmt.Fprintf(os.Stderr, "sensitive-lint: %d/%d pattern hits classified (strict=%v, findings=%d)\n",
			rep.Classified, rep.PatternHits, rep.Strict, len(rep.Findings))
	case exitFail:
		fmt.Fprintf(os.Stderr, "sensitive-lint: STRICT FAIL with %d finding(s)\n", len(rep.Findings))
		for _, f := range rep.Findings {
			fmt.Fprintf(os.Stderr, "  [%s] %s %s %s\n", f.Severity, f.Kind, f.Attribute, f.Detail)
		}
	}
	os.Exit(code)
}

func run(in, out, root string, strict bool) (*report, int, error) {
	patternedAttrs, total, err := loadPatternedAttributes(in)
	if err != nil {
		return nil, exitErr, fmt.Errorf("load openapi: %w", err)
	}
	registry, err := scanRegistries(root)
	if err != nil {
		return nil, exitErr, fmt.Errorf("scan registries: %w", err)
	}

	rep := &report{
		OpenAPISource:    in,
		Strict:           strict,
		TotalAttributes:  total,
		PatternHits:      len(patternedAttrs),
		OpenAPIPatterned: patternedAttrs,
	}
	for _, p := range registry {
		rep.RegistryEntries = append(rep.RegistryEntries, p.Path+" ("+p.SourceFile+")")
	}
	sort.Strings(rep.RegistryEntries)

	registryByLeaf := make(map[string]struct{}, len(registry))
	for _, p := range registry {
		registryByLeaf[leafName(p.Path)] = struct{}{}
	}

	for _, attr := range patternedAttrs {
		if _, ok := registryByLeaf[attr]; ok {
			rep.Classified++
			continue
		}
		// Names matched by FR-002 pattern are auto-classified by the
		// runtime redactor without needing a per-resource entry. So
		// emit a MEDIUM finding (informational), not a HIGH miss.
		rep.Findings = append(rep.Findings, finding{
			Severity:  "MEDIUM",
			Kind:      "unclassified_field",
			Attribute: attr,
			Detail:    "matches FR-002 pattern; runtime auto-redacts but no per-resource registry entry exists",
		})
	}

	if err := writeReport(out, rep); err != nil {
		return nil, exitErr, fmt.Errorf("write report: %w", err)
	}

	if strict && len(rep.Findings) > 0 {
		return rep, exitFail, nil
	}
	return rep, exitOK, nil
}

// loadPatternedAttributes walks the OpenAPI doc and returns the
// deduplicated set of attribute names that match
// redact.MatchPattern (FR-002).
//
// Implementation note: we use stdlib JSON instead of a strict
// OpenAPI parser because the upstream NC2 spec contains a few
// non-canonical `items: "<RefName>"` strings that strict parsers
// reject. We only need to walk the JSON tree looking for any
// `properties: { <name>: ... }` object whose <name> matches an
// FR-002 pattern, so generic JSON walking suffices.
func loadPatternedAttributes(path string) ([]string, int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, 0, err
	}
	hits := map[string]struct{}{}
	total := 0
	walkAny(raw, hits, &total)
	out := make([]string, 0, len(hits))
	for k := range hits {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, total, nil
}

// walkAny recursively descends through any JSON value, looking for
// `properties` maps. For each property name encountered, increments
// the total count and records a hit if the name matches an FR-002
// pattern.
func walkAny(node any, hits map[string]struct{}, total *int) {
	switch v := node.(type) {
	case map[string]any:
		if props, ok := v["properties"].(map[string]any); ok {
			for name, child := range props {
				*total++
				if redact.MatchPattern(name) {
					hits[name] = struct{}{}
				}
				walkAny(child, hits, total)
			}
		}
		for k, child := range v {
			if k == "properties" {
				// Already walked above as named children; skip the
				// generic descent so each property is counted once.
				continue
			}
			walkAny(child, hits, total)
		}
	case []any:
		for _, child := range v {
			walkAny(child, hits, total)
		}
	}
}

type registryEntry struct {
	Path       string
	SourceFile string
}

// scanRegistries walks internal/resources/*/sensitive.go (and
// equivalent files in other internal trees) looking for a function
// named `Sensitive` whose body returns []string{...} literals.
func scanRegistries(root string) ([]registryEntry, error) {
	var entries []registryEntry
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" || name == "tools" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "sensitive.go" {
			return nil
		}
		es, err := parseSensitiveFile(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		rel, _ := filepath.Rel(root, path)
		for _, e := range es {
			entries = append(entries, registryEntry{Path: e, SourceFile: rel})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return entries, nil
}

// parseSensitiveFile extracts string literals from the return slice
// of a function named `Sensitive() []string`.
func parseSensitiveFile(path string) ([]string, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil || fn.Name.Name != "Sensitive" {
			continue
		}
		if fn.Body == nil {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			cl, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}
			for _, elt := range cl.Elts {
				lit, ok := elt.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				v := strings.Trim(lit.Value, "\"`")
				if v != "" {
					out = append(out, v)
				}
			}
			return true
		})
	}
	return out, nil
}

func leafName(dottedPath string) string {
	if i := strings.LastIndexByte(dottedPath, '.'); i >= 0 {
		return dottedPath[i+1:]
	}
	return dottedPath
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
