package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type resourceAttributeContract struct {
	kind      string
	required  bool
	optional  bool
	computed  bool
	sensitive bool
}

var (
	requiredString          = resourceAttributeContract{kind: "string", required: true}
	requiredInt64           = resourceAttributeContract{kind: "int64", required: true}
	requiredSensitiveString = resourceAttributeContract{kind: "string", required: true, sensitive: true}
	optionalString          = resourceAttributeContract{kind: "string", optional: true}
	optionalMap             = resourceAttributeContract{kind: "map", optional: true}
	optionalSensitiveString = resourceAttributeContract{kind: "string", optional: true, sensitive: true}
	optionalComputedString  = resourceAttributeContract{kind: "string", optional: true, computed: true}
	optionalComputedInt64   = resourceAttributeContract{kind: "int64", optional: true, computed: true}
	optionalComputedBool    = resourceAttributeContract{kind: "bool", optional: true, computed: true}
	optionalComputedList    = resourceAttributeContract{kind: "list", optional: true, computed: true}
	computedString          = resourceAttributeContract{kind: "string", computed: true}
	computedInt64           = resourceAttributeContract{kind: "int64", computed: true}
)

func assertResourceSchemaContract(
	t *testing.T,
	r resource.Resource,
	wantAttributes map[string]resourceAttributeContract,
	wantBlocks map[string][]string,
) {
	t.Helper()

	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() returned diagnostics errors: %v", resp.Diagnostics)
	}

	if len(resp.Schema.Attributes) != len(wantAttributes) {
		t.Errorf("Schema() returned %d attributes; want %d", len(resp.Schema.Attributes), len(wantAttributes))
	}
	for name, want := range wantAttributes {
		got, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Errorf("Schema() missing attribute %q", name)
			continue
		}

		if kind := resourceAttributeKind(got); kind != want.kind {
			t.Errorf("attribute %q kind = %q; want %q", name, kind, want.kind)
		}
		if got.IsRequired() != want.required {
			t.Errorf("attribute %q required = %t; want %t", name, got.IsRequired(), want.required)
		}
		if got.IsOptional() != want.optional {
			t.Errorf("attribute %q optional = %t; want %t", name, got.IsOptional(), want.optional)
		}
		if got.IsComputed() != want.computed {
			t.Errorf("attribute %q computed = %t; want %t", name, got.IsComputed(), want.computed)
		}
		if got.IsSensitive() != want.sensitive {
			t.Errorf("attribute %q sensitive = %t; want %t", name, got.IsSensitive(), want.sensitive)
		}
	}

	if len(resp.Schema.Blocks) != len(wantBlocks) {
		t.Errorf("Schema() returned %d blocks; want %d", len(resp.Schema.Blocks), len(wantBlocks))
	}
	for name, wantNestedAttributes := range wantBlocks {
		got, ok := resp.Schema.Blocks[name]
		if !ok {
			t.Errorf("Schema() missing block %q", name)
			continue
		}

		single, ok := got.(schema.SingleNestedBlock)
		if !ok {
			t.Errorf("block %q type = %T; want schema.SingleNestedBlock", name, got)
			continue
		}
		if len(single.Attributes) != len(wantNestedAttributes) {
			t.Errorf("block %q returned %d attributes; want %d", name, len(single.Attributes), len(wantNestedAttributes))
		}
		for _, nestedName := range wantNestedAttributes {
			if _, ok := single.Attributes[nestedName]; !ok {
				t.Errorf("block %q missing attribute %q", name, nestedName)
			}
		}
	}
}

func resourceAttributeKind(attribute schema.Attribute) string {
	switch attribute.(type) {
	case schema.BoolAttribute:
		return "bool"
	case schema.Int64Attribute:
		return "int64"
	case schema.ListAttribute:
		return "list"
	case schema.MapAttribute:
		return "map"
	case schema.StringAttribute:
		return "string"
	default:
		return fmt.Sprintf("%T", attribute)
	}
}
