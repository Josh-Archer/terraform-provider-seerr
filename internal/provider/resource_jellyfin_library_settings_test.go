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

func TestJellyfinLibrarySettingsParseResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/settings/jellyfin/library", r.URL.Path)
		if r.URL.Query().Get("enable") != "" {
			assert.Equal(t, "jf1", r.URL.Query().Get("enable"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[
			{"id": "jf1", "name": "4K Movies", "enabled": true},
			{"id": "jf2", "name": "Anime", "enabled": false}
		]`))
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &JellyfinLibrarySettingsResource{client: client}

	ctx := context.Background()
	var data JellyfinLibrarySettingsModel
	data.EnabledLibraries, _ = types.SetValueFrom(ctx, types.StringType, []string{"jf1"})

	err = res.updateJellyfinLibraries(ctx, &data)
	require.NoError(t, err)

	assert.False(t, data.Libraries.IsNull())
	assert.False(t, data.EnabledLibraries.IsNull())

	var enabled []string
	diags := data.EnabledLibraries.ElementsAs(ctx, &enabled, false)
	require.False(t, diags.HasError())
	assert.Equal(t, []string{"jf1"}, enabled)
}
