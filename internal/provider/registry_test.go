package provider

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestProviderRegistrationCompleteness(t *testing.T) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatalf("Failed to parse package: %v", err)
	}

	expectedResources := make(map[string]bool)
	expectedDataSources := make(map[string]bool)

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				funcDecl, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}
				name := funcDecl.Name.Name
				if strings.HasPrefix(name, "New") && strings.HasSuffix(name, "Resource") {
					expectedResources[name] = true
				}
				if strings.HasPrefix(name, "New") && strings.HasSuffix(name, "DataSource") {
					expectedDataSources[name] = true
				}
			}
		}
	}

	actualResources := getFuncNames(RegisteredResources)
	actualDataSources := getFuncNames(RegisteredDataSources)

	for expected := range expectedResources {
		if !actualResources[expected] {
			t.Errorf("Resource %s is implemented but not registered in RegisteredResources", expected)
		}
	}

	for expected := range expectedDataSources {
		if !actualDataSources[expected] {
			t.Errorf("DataSource %s is implemented but not registered in RegisteredDataSources", expected)
		}
	}
}

func getFuncNames[T func() resource.Resource | func() datasource.DataSource](funcs []T) map[string]bool {
	names := make(map[string]bool)
	for _, f := range funcs {
		fullName := runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
		// full name includes package path, e.g., github.com/Josh-Archer/terraform-provider-seerr/internal/provider.NewAPIObjectResource
		parts := strings.Split(fullName, ".")
		name := parts[len(parts)-1]
		names[name] = true
	}
	return names
}
