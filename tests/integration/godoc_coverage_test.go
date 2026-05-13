package integration

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode"
)

// TestGodocCoverage_ExportedSymbols pins SC-007: every exported
// top-level symbol (Func, Type, Var, Const) under internal/ MUST
// have a doc comment.
//
// We don't shell out to `revive` (which would be the production
// CI gate); we re-implement the same `exported` check with the
// stdlib AST so the test can run as part of `go test ./...`.
func TestGodocCoverage_ExportedSymbols(t *testing.T) {
	t.Parallel()

	root := repoRoot(t)

	var missing []string
	walkErr := filepath.WalkDir(filepath.Join(root, "internal"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, perr := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if perr != nil {
			return perr
		}
		rel, _ := filepath.Rel(root, path)
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if !isExportedName(d.Name.Name) {
					continue
				}
				// Methods on unexported types do not need a doc per
				// revive's exported rule.
				if d.Recv != nil && !receiverIsExported(d.Recv) {
					continue
				}
				if d.Doc == nil || strings.TrimSpace(d.Doc.Text()) == "" {
					missing = append(missing, rel+":"+d.Name.Name)
				}
			case *ast.GenDecl:
				switch d.Tok {
				case token.TYPE, token.VAR, token.CONST:
					for _, spec := range d.Specs {
						checkSpec(spec, d, rel, &missing)
					}
				}
			}
		}
		return nil
	})
	if walkErr != nil {
		t.Fatalf("walk: %v", walkErr)
	}
	if len(missing) > 0 {
		t.Errorf("SC-007 violation: %d exported symbol(s) missing doc comment:\n  %s",
			len(missing), strings.Join(missing, "\n  "))
	}
}

func checkSpec(spec ast.Spec, parent *ast.GenDecl, rel string, missing *[]string) {
	groupHasDoc := parent.Doc != nil && strings.TrimSpace(parent.Doc.Text()) != ""
	switch s := spec.(type) {
	case *ast.TypeSpec:
		if !isExportedName(s.Name.Name) {
			return
		}
		if (s.Doc == nil || strings.TrimSpace(s.Doc.Text()) == "") && !groupHasDoc {
			*missing = append(*missing, rel+":type "+s.Name.Name)
		}
	case *ast.ValueSpec:
		for _, name := range s.Names {
			if !isExportedName(name.Name) {
				continue
			}
			specHasDoc := s.Doc != nil && strings.TrimSpace(s.Doc.Text()) != ""
			specHasComment := s.Comment != nil && strings.TrimSpace(s.Comment.Text()) != ""
			if specHasDoc || specHasComment || groupHasDoc {
				continue
			}
			*missing = append(*missing, rel+":var/const "+name.Name)
		}
	}
}

func isExportedName(name string) bool {
	for _, r := range name {
		return unicode.IsUpper(r)
	}
	return false
}

// receiverIsExported reports whether the receiver type of a method
// is exported (i.e. starts with an uppercase letter, possibly via
// pointer wrapping).
func receiverIsExported(recv *ast.FieldList) bool {
	if recv == nil || len(recv.List) == 0 {
		return false
	}
	expr := recv.List[0].Type
	for {
		switch e := expr.(type) {
		case *ast.StarExpr:
			expr = e.X
		case *ast.Ident:
			return isExportedName(e.Name)
		case *ast.IndexExpr:
			expr = e.X
		case *ast.IndexListExpr:
			expr = e.X
		default:
			return false
		}
	}
}

// stat helper used by the loader.
func init() { _ = os.Stat }
