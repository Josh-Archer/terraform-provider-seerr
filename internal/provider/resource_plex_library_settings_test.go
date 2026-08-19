package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlexLibrarySettingsParseResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/settings/plex/library", r.URL.Path)
		if r.URL.Query().Get("enable") != "" {
			assert.Equal(t, "1,2", r.URL.Query().Get("enable"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id": "1", "name": "Movies", "enabled": true},
			{"id": "2", "name": "TV Shows", "enabled": true},
			{"id": "3", "name": "Music", "enabled": false}
		]`))
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &PlexLibrarySettingsResource{client: client}

	ctx := context.Background()
	var data PlexLibrarySettingsModel
	data.EnabledLibraries, _ = types.SetValueFrom(ctx, types.StringType, []string{"1", "2"})

	err = res.updatePlexLibraries(ctx, &data)
	require.NoError(t, err)

	assert.False(t, data.Libraries.IsNull())
	assert.False(t, data.EnabledLibraries.IsNull())

	var enabled []string
	diags := data.EnabledLibraries.ElementsAs(ctx, &enabled, false)
	require.False(t, diags.HasError())
	assert.ElementsMatch(t, []string{"1", "2"}, enabled)
}

func TestPlexLibrarySettingsEmptyEnabledLibraries(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/settings/plex/library", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("enable"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &PlexLibrarySettingsResource{client: client}

	ctx := context.Background()
	var data PlexLibrarySettingsModel
	data.EnabledLibraries, _ = types.SetValueFrom(ctx, types.StringType, []string{})

	err = res.updatePlexLibraries(ctx, &data)
	require.NoError(t, err)
}
