package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbyLibrarySettingsResourceUpdateAndRead(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/emby/library" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"id": "1", "name": "Movies", "enabled": true},
				{"id": "2", "name": "TV Shows", "enabled": false}
			]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &EmbyLibrarySettingsResource{client: client}

	ctx := context.Background()
	model := &EmbyLibrarySettingsModel{
		SyncOnRead: types.BoolValue(false),
	}

	err = res.readEmbyLibraries(ctx, model)
	require.NoError(t, err)

	assert.Equal(t, 2, len(model.Libraries.Elements()))
	assert.Equal(t, 1, len(model.EnabledLibraries.Elements()))
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

	// User imports with arbitrary string or ID per docs: `terraform import seerr_emby_library_settings.this 1`
	req := resource.ImportStateRequest{ID: "1"}
	importResp := resource.ImportStateResponse{State: state}
	r.ImportState(context.Background(), req, &importResp)
	require.False(t, importResp.Diagnostics.HasError())

	var importedData EmbyLibrarySettingsModel
	require.False(t, importResp.State.Get(context.Background(), &importedData).HasError())

	// The imported ID must immediately be the canonical singleton ID "emby_library_settings"
	// so that subsequent Read() calls do not drift or mutate the state's ID.
	assert.Equal(t, "emby_library_settings", importedData.ID.ValueString(), "Import should normalize ID to canonical singleton value 'emby_library_settings'")
}
