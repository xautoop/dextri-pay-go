package architecture

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

type packagePolicy struct {
	mayImportInternal bool
}

var publicPackagePolicies = map[string]packagePolicy{
	"api":        {},
	"channels":   {},
	"checkout":   {},
	"client":     {mayImportInternal: true},
	"conversion": {},
	"money":      {},
	"operation":  {},
	"users":      {},
	"webhook":    {},
}

func TestRepositoryRootContainsNoGoSource(t *testing.T) {
	root := repositoryRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("repository root must contain no Go source, found %v", files)
	}
}

func TestDomainPackagesDoNotImportInternalImplementation(t *testing.T) {
	root := repositoryRoot(t)
	for directory, policy := range publicPackagePolicies {
		if policy.mayImportInternal {
			continue
		}
		path := filepath.Join(root, directory)
		entries, err := os.ReadDir(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}
			filePath := filepath.Join(path, entry.Name())
			parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, imported := range parsed.Imports {
				value, err := strconv.Unquote(imported.Path.Value)
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(value, "/internal/") {
					t.Errorf("public domain package %s imports implementation package %s", directory, value)
				}
			}
		}
	}
}

func TestPublicPackagePolicyCoversTopLevelGoPackages(t *testing.T) {
	root := repositoryRoot(t)
	directories, err := os.ReadDir(root)
	if err != nil {
		t.Fatal(err)
	}
	var missing []string
	for _, directory := range directories {
		if !directory.IsDir() || directory.Name() == "internal" || directory.Name() == "examples" {
			continue
		}
		matches, err := filepath.Glob(filepath.Join(root, directory.Name(), "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) == 0 {
			continue
		}
		if _, ok := publicPackagePolicies[directory.Name()]; !ok {
			missing = append(missing, directory.Name())
		}
	}
	sort.Strings(missing)
	if len(missing) != 0 {
		t.Fatalf("top-level Go packages require an explicit architecture policy: %v", missing)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate architecture test")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}
