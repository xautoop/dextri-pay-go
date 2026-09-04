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
	"account":    {},
	"api":        {},
	"channels":   {},
	"checkout":   {},
	"client":     {mayImportInternal: true},
	"conversion": {},
	"escrow":     {},
	"hold":       {},
	"money":      {},
	"operation":  {},
	"payment":    {},
	"payout":     {},
	"tron":       {},
	"users":      {},
	"webhook":    {},
}

var domainImportPaths = map[string]struct{}{
	"github.com/xautoop/dextri-pay-go/account":    {},
	"github.com/xautoop/dextri-pay-go/channels":   {},
	"github.com/xautoop/dextri-pay-go/checkout":   {},
	"github.com/xautoop/dextri-pay-go/conversion": {},
	"github.com/xautoop/dextri-pay-go/escrow":     {},
	"github.com/xautoop/dextri-pay-go/hold":       {},
	"github.com/xautoop/dextri-pay-go/operation":  {},
	"github.com/xautoop/dextri-pay-go/payment":    {},
	"github.com/xautoop/dextri-pay-go/payout":     {},
	"github.com/xautoop/dextri-pay-go/tron":       {},
	"github.com/xautoop/dextri-pay-go/users":      {},
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

func TestLegacyResourceLayerDoesNotReturn(t *testing.T) {
	path := filepath.Join(repositoryRoot(t), "internal", "resource")
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("legacy forwarding layer must not exist: %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatal(err)
	}
}

func TestTransportDoesNotImportDomainPackages(t *testing.T) {
	root := repositoryRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "internal", "transport", "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, filePath := range files {
		parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range parsed.Imports {
			value, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if _, forbidden := domainImportPaths[value]; forbidden {
				t.Errorf("transport package imports domain package %s in %s", value, filePath)
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
