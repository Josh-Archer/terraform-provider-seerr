package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJellyfinLibrarySettingsResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewJellyfinLibrarySettingsResource(), map[string]resourceAttributeContract{
		"id":                computedString,
		"sync_on_read":      optionalComputedBoolKeepState,
		"enabled_libraries": requiredSet,
		"libraries":         computedListNestedKeepState,
	}, nil)
}

func TestJellyfinLibrarySettingsCreateThenReadPersistsEnabledLibraries(t *testing.T) {
	srv := newLibrarySettingsServer(t, "/api/v1/settings/jellyfin/library", []mockLibrary{
		{ID: "jf1", Name: "Movies", Enabled: false},
		{ID: "jf2", Name: "TV", Enabled: false},
	})
	res := &JellyfinLibrarySettingsResource{client: testAPIClient(t, srv.URL)}
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)

	enabled, diags := types.SetValueFrom(ctx, types.StringType, []string{"jf1"})
	require.False(t, diags.HasError())
	planModel := JellyfinLibrarySettingsModel{
		EnabledLibraries: enabled,
		Libraries:        types.ListNull(libraryObjectType()),
	}
	plan := tfsdk.Plan{Schema: schemaResp.Schema}
	require.False(t, plan.Set(ctx, &planModel).HasError())

	createResp := resource.CreateResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	res.Create(ctx, resource.CreateRequest{Plan: plan}, &createResp)
	require.False(t, createResp.Diagnostics.HasError(), createResp.Diagnostics)

	var afterCreate JellyfinLibrarySettingsModel
	require.False(t, createResp.State.Get(ctx, &afterCreate).HasError())
	assertEnabledLibraryIDs(t, ctx, afterCreate.EnabledLibraries, []string{"jf1"})
	assert.False(t, afterCreate.SyncOnRead.ValueBool())
	assert.Equal(t, "jellyfin_library_settings", afterCreate.ID.ValueString())

	readResp := resource.ReadResponse{State: createResp.State}
	res.Read(ctx, resource.ReadRequest{State: createResp.State}, &readResp)
	require.False(t, readResp.Diagnostics.HasError(), readResp.Diagnostics)

	var afterRead JellyfinLibrarySettingsModel
	require.False(t, readResp.State.Get(ctx, &afterRead).HasError())
	assertEnabledLibraryIDs(t, ctx, afterRead.EnabledLibraries, []string{"jf1"})
	assert.Equal(t, afterCreate.Libraries, afterRead.Libraries)
	assert.Equal(t, afterCreate.SyncOnRead, afterRead.SyncOnRead)
}

func libraryObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":      types.StringType,
			"name":    types.StringType,
			"enabled": types.BoolType,
		},
	}
}

func assertEnabledLibraryIDs(t *testing.T, ctx context.Context, got types.Set, want []string) {
	t.Helper()
	var ids []string
	diags := got.ElementsAs(ctx, &ids, false)
	require.False(t, diags.HasError())
	assert.ElementsMatch(t, want, ids)
}
