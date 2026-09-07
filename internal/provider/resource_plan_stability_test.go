package provider

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestComputedCollectionsKeepStateForUnknown(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var violations []string
	for _, factory := range RegisteredResources {
		r := factory()
		name := resourceTypeName(ctx, r)
		for _, v := range computedCollectionStabilityViolations(ctx, r) {
			violations = append(violations, name+"."+v)
		}
	}
	if len(violations) > 0 {
		t.Fatalf("computed collection attributes missing UseStateForUnknown (perpetual plan):\n  %s", strings.Join(violations, "\n  "))
	}
}

func TestOptionalComputedBoolsStabilize(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	var violations []string
	for _, factory := range RegisteredResources {
		r := factory()
		name := resourceTypeName(ctx, r)
		for _, v := range optionalComputedBoolStabilityViolations(ctx, r) {
			violations = append(violations, name+"."+v)
		}
	}
	if len(violations) > 0 {
		t.Fatalf("optional+computed bools missing Default and UseStateForUnknown (perpetual plan):\n  %s", strings.Join(violations, "\n  "))
	}
}

func TestResourcesDoNotMutateViaGETQuery(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	fset := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		if strings.HasPrefix(file, "data_source_") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if !strings.Contains(string(src), "enable=") {
			continue
		}
		node, err := parser.ParseFile(fset, file, src, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		ast.Inspect(node, func(n ast.Node) bool {
			bl, ok := n.(*ast.BasicLit)
			if !ok {
				return true
			}
			if strings.Contains(bl.Value, "enable=") {
				pos := fset.Position(bl.Pos())
				t.Errorf("%s:%d: resource write path must not encode enablement as GET ?enable=; use PUT /library/{id}", file, pos.Line)
			}
			return true
		})
	}
}

func resourceTypeName(ctx context.Context, r resource.Resource) string {
	var meta resource.MetadataResponse
	r.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "seerr"}, &meta)
	return meta.TypeName
}

func computedCollectionStabilityViolations(ctx context.Context, r resource.Resource) []string {
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	var out []string
	walkResourceSchema(schemaResp.Schema, func(path string, attr schema.Attribute) {
		if !isComputedCollection(attr) {
			return
		}
		if attr.IsOptional() || attr.IsRequired() {
			return
		}
		if !attributeKeepsStateForUnknown(attr) {
			out = append(out, path)
		}
	})
	return out
}

func optionalComputedBoolStabilityViolations(ctx context.Context, r resource.Resource) []string {
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	var out []string
	walkResourceSchema(schemaResp.Schema, func(path string, attr schema.Attribute) {
		boolAttr, ok := attr.(schema.BoolAttribute)
		if !ok {
			return
		}
		if !attr.IsOptional() || !attr.IsComputed() {
			return
		}
		if boolAttr.Default != nil || attributeKeepsStateForUnknown(attr) {
			return
		}
		out = append(out, path)
	})
	return out
}

func isComputedCollection(attr schema.Attribute) bool {
	switch attr.(type) {
	case schema.ListAttribute, schema.ListNestedAttribute, schema.SetAttribute, schema.SetNestedAttribute, schema.MapAttribute, schema.MapNestedAttribute:
		return attr.IsComputed()
	default:
		return false
	}
}

func walkResourceSchema(s schema.Schema, fn func(path string, attr schema.Attribute)) {
	walkSchemaAttributes(s.Attributes, "", fn)
	walkSchemaBlocks(s.Blocks, "", fn)
}

func walkSchemaBlocks(blocks map[string]schema.Block, prefix string, fn func(path string, attr schema.Attribute)) {
	for name, block := range blocks {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		switch b := block.(type) {
		case schema.SingleNestedBlock:
			walkSchemaAttributes(b.Attributes, path, fn)
			walkSchemaBlocks(b.Blocks, path, fn)
		case schema.ListNestedBlock:
			walkSchemaAttributes(b.NestedObject.Attributes, path, fn)
			walkSchemaBlocks(b.NestedObject.Blocks, path, fn)
		case schema.SetNestedBlock:
			walkSchemaAttributes(b.NestedObject.Attributes, path, fn)
			walkSchemaBlocks(b.NestedObject.Blocks, path, fn)
		}
	}
}

func walkSchemaAttributes(attrs map[string]schema.Attribute, prefix string, fn func(path string, attr schema.Attribute)) {
	for name, attr := range attrs {
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		fn(path, attr)
		switch nested := attr.(type) {
		case schema.ListNestedAttribute:
			walkSchemaAttributes(nested.NestedObject.Attributes, path, fn)
		case schema.SetNestedAttribute:
			walkSchemaAttributes(nested.NestedObject.Attributes, path, fn)
		case schema.SingleNestedAttribute:
			walkSchemaAttributes(nested.Attributes, path, fn)
		case schema.MapNestedAttribute:
			walkSchemaAttributes(nested.NestedObject.Attributes, path, fn)
		}
	}
}

func TestCoveredLibraryWriteEndpointsUsePUT(t *testing.T) {
	t.Parallel()
	root := repoRootFromTest(t)
	src, err := os.ReadFile(filepath.Join(root, "internal", "provider", "library_settings.go"))
	if err != nil {
		t.Fatalf("read library_settings.go: %v", err)
	}
	body := string(src)
	if !strings.Contains(body, "http.MethodPut") {
		t.Fatal("library enablement must PUT /{libraryId}; list GET is read-only")
	}
	if strings.Contains(body, "?enable=") {
		t.Fatal("library enablement still encodes a write as a list GET query")
	}
}

func TestOpenAPILibraryWritePathsAreImplemented(t *testing.T) {
	t.Parallel()
	files := []string{
		"library_settings.go",
		"resource_jellyfin_library_settings.go",
		"resource_plex_library_settings.go",
		"resource_emby_library_settings.go",
	}
	joined := ""
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		joined += string(src)
	}
	for _, media := range []string{"jellyfin", "plex", "emby"} {
		base := fmt.Sprintf("/api/v1/settings/%s/library", media)
		if !strings.Contains(joined, base) {
			t.Errorf("missing library settings path %s", base)
		}
	}
	if !strings.Contains(joined, "http.MethodPut") {
		t.Error("library settings write path must use PUT")
	}
}
