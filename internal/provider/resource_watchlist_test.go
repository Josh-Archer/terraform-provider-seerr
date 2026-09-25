package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchlistResource_Schema(t *testing.T) {
	r := NewWatchlistResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics error: %v", resp.Diagnostics)
	}

	attrs := []string{"id", "tmdb_id", "media_type", "title", "overview"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Expected '%s' attribute in schema", attr)
		}
	}
}

func TestWatchlistResource_Metadata(t *testing.T) {
	r := NewWatchlistResource()
	var metaResp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_watchlist" {
		t.Errorf("Expected type name seerr_watchlist, got %s", metaResp.TypeName)
	}
}

func TestWatchlistResourceSchemaRequiresReplace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := NewWatchlistResource()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())

	nonNullState := tfsdk.State{Raw: tftypes.NewValue(tftypes.String, "state")}
	nonNullPlan := tfsdk.Plan{Raw: tftypes.NewValue(tftypes.String, "plan")}

	// tmdb_id requires replace
	tmdbAttr, ok := resp.Schema.Attributes["tmdb_id"].(schema.Int64Attribute)
	require.True(t, ok, "tmdb_id attribute not found or not Int64Attribute")
	require.NotEmpty(t, tmdbAttr.PlanModifiers, "tmdb_id missing plan modifiers")

	hasInt64RequiresReplace := false
	for _, mod := range tmdbAttr.PlanModifiers {
		if strings.Contains(mod.Description(ctx), "destroy and recreate") {
			hasInt64RequiresReplace = true
		}
	}
	assert.True(t, hasInt64RequiresReplace, "tmdb_id must have RequiresReplace plan modifier")

	int64Req := planmodifier.Int64Request{
		State:      nonNullState,
		Plan:       nonNullPlan,
		StateValue: types.Int64Value(123),
		PlanValue:  types.Int64Value(456),
	}
	int64Resp := planmodifier.Int64Response{}
	for _, mod := range tmdbAttr.PlanModifiers {
		mod.PlanModifyInt64(ctx, int64Req, &int64Resp)
	}
	assert.True(t, int64Resp.RequiresReplace, "tmdb_id modification should require replacement")

	int64ReqSame := planmodifier.Int64Request{
		State:      nonNullState,
		Plan:       nonNullPlan,
		StateValue: types.Int64Value(123),
		PlanValue:  types.Int64Value(123),
	}
	int64RespSame := planmodifier.Int64Response{}
	for _, mod := range tmdbAttr.PlanModifiers {
		mod.PlanModifyInt64(ctx, int64ReqSame, &int64RespSame)
	}
	assert.False(t, int64RespSame.RequiresReplace, "unchanged tmdb_id should not require replacement")

	// media_type requires replace
	mediaTypeAttr, ok := resp.Schema.Attributes["media_type"].(schema.StringAttribute)
	require.True(t, ok, "media_type attribute not found or not StringAttribute")
	require.NotEmpty(t, mediaTypeAttr.PlanModifiers, "media_type missing plan modifiers")

	hasStringRequiresReplace := false
	for _, mod := range mediaTypeAttr.PlanModifiers {
		if strings.Contains(mod.Description(ctx), "destroy and recreate") {
			hasStringRequiresReplace = true
		}
	}
	assert.True(t, hasStringRequiresReplace, "media_type must have RequiresReplace plan modifier")

	stringReq := planmodifier.StringRequest{
		State:      nonNullState,
		Plan:       nonNullPlan,
		StateValue: types.StringValue("movie"),
		PlanValue:  types.StringValue("tv"),
	}
	stringResp := planmodifier.StringResponse{}
	for _, mod := range mediaTypeAttr.PlanModifiers {
		mod.PlanModifyString(ctx, stringReq, &stringResp)
	}
	assert.True(t, stringResp.RequiresReplace, "media_type modification should require replacement")

	stringReqSame := planmodifier.StringRequest{
		State:      nonNullState,
		Plan:       nonNullPlan,
		StateValue: types.StringValue("movie"),
		PlanValue:  types.StringValue("movie"),
	}
	stringRespSame := planmodifier.StringResponse{}
	for _, mod := range mediaTypeAttr.PlanModifiers {
		mod.PlanModifyString(ctx, stringReqSame, &stringRespSame)
	}
	assert.False(t, stringRespSame.RequiresReplace, "unchanged media_type should not require replacement")
}
