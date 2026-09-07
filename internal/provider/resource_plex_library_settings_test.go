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

func TestPlexLibrarySettingsResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewPlexLibrarySettingsResource(), map[string]resourceAttributeContract{
		"id":                computedString,
		"sync_on_read":      optionalComputedBoolKeepState,
		"enabled_libraries": requiredSet,
		"libraries":         computedListNestedKeepState,
	}, nil)
}

func TestPlexLibrarySettingsUpdateUsesEnabledLibraries(t *testing.T) {
	srv := newLibrarySettingsServer(t, "/api/v1/settings/plex/library", []mockLibrary{
		{ID: "1", Name: "Movies", Enabled: false},
		{ID: "2", Name: "TV", Enabled: false},
		{ID: "3", Name: "Music", Enabled: false},
	})
	res := &PlexLibrarySettingsResource{client: testAPIClient(t, srv.URL)}
	ctx := context.Background()

	var data PlexLibrarySettingsModel
	data.EnabledLibraries, _ = types.SetValueFrom(ctx, types.StringType, []string{"1", "2"})

	require.NoError(t, res.updatePlexLibraries(ctx, &data))
	assertEnabledLibraryIDs(t, ctx, data.EnabledLibraries, []string{"1", "2"})

	require.NoError(t, res.readPlexLibraries(ctx, &data))
	assertEnabledLibraryIDs(t, ctx, data.EnabledLibraries, []string{"1", "2"})
}

func TestPlexLibrarySettingsEmptyEnabledLibrariesDisablesAll(t *testing.T) {
	srv := newLibrarySettingsServer(t, "/api/v1/settings/plex/library", []mockLibrary{
		{ID: "1", Name: "Movies", Enabled: true},
	})
	res := &PlexLibrarySettingsResource{client: testAPIClient(t, srv.URL)}
	ctx := context.Background()

	var data PlexLibrarySettingsModel
	data.EnabledLibraries, _ = types.SetValueFrom(ctx, types.StringType, []string{})

	require.NoError(t, res.updatePlexLibraries(ctx, &data))
	assertEnabledLibraryIDs(t, ctx, data.EnabledLibraries, nil)
}

func TestPlexLibrarySettingsImportIDConsistency(t *testing.T) {
	r := &PlexLibrarySettingsResource{}
	var schemaResp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &schemaResp)

	state := tfsdk.State{Schema: schemaResp.Schema}
	initialModel := PlexLibrarySettingsModel{
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

	var importedData PlexLibrarySettingsModel
	require.False(t, importResp.State.Get(context.Background(), &importedData).HasError())
	assert.Equal(t, "plex_library_settings", importedData.ID.ValueString())
}
