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

func TestEmbyLibrarySettingsResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewEmbyLibrarySettingsResource(), map[string]resourceAttributeContract{
		"id":                computedString,
		"sync_on_read":      optionalComputedBoolKeepState,
		"enabled_libraries": requiredSet,
		"libraries":         computedListNestedKeepState,
	}, nil)
}

func TestEmbyLibrarySettingsResourceUpdateAndRead(t *testing.T) {
	srv := newLibrarySettingsServer(t, "/api/v1/settings/emby/library", []mockLibrary{
		{ID: "1", Name: "Movies", Enabled: false},
		{ID: "2", Name: "TV Shows", Enabled: false},
	})
	res := &EmbyLibrarySettingsResource{client: testAPIClient(t, srv.URL)}
	ctx := context.Background()

	model := &EmbyLibrarySettingsModel{
		SyncOnRead: types.BoolValue(false),
	}
	model.EnabledLibraries, _ = types.SetValueFrom(ctx, types.StringType, []string{"1"})

	require.NoError(t, res.updateEmbyLibraries(ctx, model))
	assertEnabledLibraryIDs(t, ctx, model.EnabledLibraries, []string{"1"})
	assert.Equal(t, 2, len(model.Libraries.Elements()))

	require.NoError(t, res.readEmbyLibraries(ctx, model))
	assertEnabledLibraryIDs(t, ctx, model.EnabledLibraries, []string{"1"})
}

func TestEmbyLibrarySettingsImportIDConsistency(t *testing.T) {
	r := &EmbyLibrarySettingsResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	state := tfsdk.State{Schema: schemaResp.Schema}
	initialModel := EmbyLibrarySettingsModel{
		EnabledLibraries: types.SetNull(types.StringType),
		Libraries: types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"id":      types.StringType,
				"name":    types.StringType,
				"enabled": types.BoolType,
			},
		}),
	}
	require.False(t, state.Set(context.Background(), &initialModel).HasError())

	req := resource.ImportStateRequest{ID: "1"}
	importResp := resource.ImportStateResponse{State: state}
	r.ImportState(context.Background(), req, &importResp)
	require.False(t, importResp.Diagnostics.HasError())

	var importedData EmbyLibrarySettingsModel
	require.False(t, importResp.State.Get(context.Background(), &importedData).HasError())
	assert.Equal(t, "emby_library_settings", importedData.ID.ValueString(), "Import should normalize ID to canonical singleton value 'emby_library_settings'")
}
