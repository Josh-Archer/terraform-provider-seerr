package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

func TestBlocklistResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewBlocklistResource(), map[string]resourceAttributeContract{
		"id":               computedString,
		"tmdb_id":          requiredInt64,
		"media_type":       requiredString,
		"title":            optionalComputedString,
		"user_id":          requiredInt64,
		"blocklisted_tags": optionalComputedString,
		"created_at":       computedString,
	}, nil)
}

func TestBlocklistResourceImmutableAttributesRequireReplacement(t *testing.T) {
	var resp resource.SchemaResponse
	(&BlocklistResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)

	for _, name := range []string{"tmdb_id", "media_type", "user_id"} {
		attr, ok := resp.Schema.Attributes[name]
		if !ok {
			t.Fatalf("schema missing %q", name)
		}

		var modifierCount int
		switch typed := attr.(type) {
		case rschema.Int64Attribute:
			modifierCount = len(typed.PlanModifiers)
		case rschema.StringAttribute:
			modifierCount = len(typed.PlanModifiers)
		default:
			t.Fatalf("attribute %q has unexpected type %T", name, attr)
		}
		if modifierCount == 0 {
			t.Errorf("attribute %q must require replacement; got no plan modifiers (%s)", name, fmt.Sprintf("%T", attr))
		}
	}
}
