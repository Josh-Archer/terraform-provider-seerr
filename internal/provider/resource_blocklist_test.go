package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/require"
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

func TestBlocklistResourceSchemaRequiresReplace(t *testing.T) {
	t.Parallel()
	r := NewBlocklistResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())

	for _, attrName := range []string{"title", "blocklisted_tags", "tmdb_id", "media_type", "user_id"} {
		attr, ok := resp.Schema.Attributes[attrName]
		require.True(t, ok, "attribute %s not found", attrName)
		switch a := attr.(type) {
		case rschema.StringAttribute:
			require.NotEmpty(t, a.PlanModifiers, "string attribute %s missing plan modifiers", attrName)
		case rschema.Int64Attribute:
			require.NotEmpty(t, a.PlanModifiers, "int64 attribute %s missing plan modifiers", attrName)
		}
	}
}
