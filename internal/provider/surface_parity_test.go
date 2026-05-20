package provider

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProviderSurfaceParity(t *testing.T) {
	repoRoot := repoRootFromTest(t)
	provider := New("test")()

	resourceNames := registeredResourceNames(t, provider.Resources(context.Background()))
	dataSourceNames := registeredDataSourceNames(t, provider.DataSources(context.Background()))

	for name := range resourceNames {
		doc := filepath.Join(repoRoot, "docs", "resources", trimProviderPrefix(name)+".md")
		if _, err := os.Stat(doc); err != nil {
			t.Fatalf("missing resource doc for %s at %s", name, doc)
		}
	}
	for name := range dataSourceNames {
		doc := filepath.Join(repoRoot, "docs", "data-sources", trimProviderPrefix(name)+".md")
		if _, err := os.Stat(doc); err != nil {
			t.Fatalf("missing data source doc for %s at %s", name, doc)
		}
	}

	assertExampleDirectoriesRegistered(t, filepath.Join(repoRoot, "examples", "resources"), resourceNames)
	assertExampleDirectoriesRegistered(t, filepath.Join(repoRoot, "examples", "data-sources"), dataSourceNames)
	assertModulesReferenceRegisteredResources(t, filepath.Join(repoRoot, "modules"), resourceNames)
	assertModulesUseRegisteredResourceAttributes(t, filepath.Join(repoRoot, "modules"), provider.Resources(context.Background()))
	assertLocalModuleSourcesExist(t, filepath.Join(repoRoot, "examples", "modules"))
}

func repoRootFromTest(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to determine caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func registeredResourceNames(t *testing.T, factories []func() resource.Resource) map[string]struct{} {
	t.Helper()
	names := make(map[string]struct{}, len(factories))
	for _, factory := range factories {
		res := factory()
		var resp resource.MetadataResponse
		res.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "seerr"}, &resp)
		names[resp.TypeName] = struct{}{}
	}
	return names
}

func registeredDataSourceNames(t *testing.T, factories []func() datasource.DataSource) map[string]struct{} {
	t.Helper()
	names := make(map[string]struct{}, len(factories))
	for _, factory := range factories {
		ds := factory()
		var resp datasource.MetadataResponse
		ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, &resp)
		names[resp.TypeName] = struct{}{}
	}
	return names
}

func assertExampleDirectoriesRegistered(t *testing.T, base string, registered map[string]struct{}) {
	t.Helper()
	entries, err := os.ReadDir(base)
	if err != nil {
		t.Fatalf("read dir %s: %v", base, err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, ok := registered[name]; !ok {
			t.Fatalf("example directory %s/%s does not match a registered provider type", base, name)
		}
	}
}

func assertModulesReferenceRegisteredResources(t *testing.T, modulesDir string, registered map[string]struct{}) {
	t.Helper()
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		t.Fatalf("read dir %s: %v", modulesDir, err)
	}

	resourceRefPattern := regexp.MustCompile(`resource\s+"([^"]+)"\s+"[^"]+"`)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		mainPath := filepath.Join(modulesDir, entry.Name(), "main.tf")
		body, err := os.ReadFile(mainPath)
		if err != nil {
			continue
		}
		matches := resourceRefPattern.FindAllStringSubmatch(string(body), -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			resourceName := strings.TrimSpace(match[1])
			if _, ok := registered[resourceName]; !ok {
				t.Fatalf("module %s references unregistered resource %s", entry.Name(), resourceName)
			}
		}
	}
}

func assertModulesUseRegisteredResourceAttributes(t *testing.T, modulesDir string, factories []func() resource.Resource) {
	t.Helper()
	attributes := registeredResourceAttributes(t, factories)
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		t.Fatalf("read dir %s: %v", modulesDir, err)
	}

	resourceBlockPattern := regexp.MustCompile(`(?s)resource\s+"([^"]+)"\s+"[^"]+"\s+\{(.*?)\n\}`)
	attributePattern := regexp.MustCompile(`(?m)^\s*([A-Za-z0-9_]+)\s*=`)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		mainPath := filepath.Join(modulesDir, entry.Name(), "main.tf")
		body, err := os.ReadFile(mainPath)
		if err != nil {
			continue
		}
		for _, block := range resourceBlockPattern.FindAllStringSubmatch(string(body), -1) {
			if len(block) < 3 {
				continue
			}
			resourceName := block[1]
			allowed, ok := attributes[resourceName]
			if !ok {
				continue
			}
			for _, match := range attributePattern.FindAllStringSubmatch(block[2], -1) {
				if len(match) < 2 {
					continue
				}
				attrName := strings.TrimSpace(match[1])
				if _, ok := allowed[attrName]; !ok {
					t.Fatalf("module %s references unsupported attribute %s.%s", entry.Name(), resourceName, attrName)
				}
			}
		}
	}
}

func registeredResourceAttributes(t *testing.T, factories []func() resource.Resource) map[string]map[string]struct{} {
	t.Helper()
	out := map[string]map[string]struct{}{}
	for _, factory := range factories {
		res := factory()
		var metadata resource.MetadataResponse
		res.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "seerr"}, &metadata)
		var schemaResp resource.SchemaResponse
		res.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)
		attrs := map[string]struct{}{}
		for name := range schemaResp.Schema.Attributes {
			attrs[name] = struct{}{}
		}
		for name := range schemaResp.Schema.Blocks {
			attrs[name] = struct{}{}
		}
		out[metadata.TypeName] = attrs
	}
	return out
}

func assertLocalModuleSourcesExist(t *testing.T, modulesDir string) {
	t.Helper()
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		t.Fatalf("read dir %s: %v", modulesDir, err)
	}
	sourcePattern := regexp.MustCompile(`source\s*=\s*"(\.\.?/[^"]+)"`)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		mainPath := filepath.Join(modulesDir, entry.Name(), "main.tf")
		body, err := os.ReadFile(mainPath)
		if err != nil {
			continue
		}
		for _, match := range sourcePattern.FindAllStringSubmatch(string(body), -1) {
			if len(match) < 2 {
				continue
			}
			target := filepath.Clean(filepath.Join(filepath.Dir(mainPath), match[1]))
			if _, err := os.Stat(target); err != nil {
				t.Fatalf("example module %s references missing local source %s", entry.Name(), match[1])
			}
		}
	}
}

func trimProviderPrefix(name string) string {
	return strings.TrimPrefix(name, "seerr_")
}
