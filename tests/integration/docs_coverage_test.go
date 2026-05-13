package integration

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDocsCoverage pins FR-029 and SC-006: every resource / data
// source / action package under internal/{resources,datasources,
// actions}/ MUST have a corresponding markdown doc under
// docs/{resources,data-sources,actions}/<name>.md AND at least one
// runnable example under examples/{resources,data-sources,actions}/<name>/.
func TestDocsCoverage(t *testing.T) {
	root := repoRoot(t)

	cases := []struct {
		kind      string // package family
		docDir    string // docs subdir
		exampleDir string // examples subdir
		excluded  map[string]bool
	}{
		{
			kind:      "resources",
			docDir:    "resources",
			exampleDir: "resources",
			excluded: map[string]bool{
				// Internal helpers not exposed as Terraform resources.
				"clustershared": true,
				"orgshared":     true,
			},
		},
		{
			kind:      "datasources",
			docDir:    "data-sources",
			exampleDir: "data-sources",
			excluded: map[string]bool{
				"dsshared": true,
			},
		},
		{
			kind:      "actions",
			docDir:    "actions",
			exampleDir: "actions",
			excluded: map[string]bool{
				"shared": true,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.kind, func(t *testing.T) {
			t.Parallel()
			pkgs, err := listSubdirs(filepath.Join(root, "internal", tc.kind))
			if err != nil {
				t.Fatalf("list internal/%s: %v", tc.kind, err)
			}
			for _, pkg := range pkgs {
				if tc.excluded[pkg] {
					continue
				}
				resourceName := "nc2_" + pkg
				docPath := filepath.Join(root, "docs", tc.docDir, resourceName+".md")
				if !fileExists(docPath) {
					t.Errorf("missing docs file: %s", docPath)
				}
				exampleParent := filepath.Join(root, "examples", tc.exampleDir, resourceName)
				if !dirHasTerraformFile(exampleParent) {
					t.Errorf("missing or empty examples dir: %s", exampleParent)
				}
			}
		})
	}
}

func listSubdirs(parent string) ([]string, error) {
	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			out = append(out, e.Name())
		}
	}
	return out, nil
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// dirHasTerraformFile returns true if `dir` contains at least one
// .tf file at any depth (catches per-cloud subdirs like
// examples/resources/nc2_cloud_account/aws/).
func dirHasTerraformFile(dir string) bool {
	if !fileExists(dir) {
		return false
	}
	found := false
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(path, ".tf") {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}
