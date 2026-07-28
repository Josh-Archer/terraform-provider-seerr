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

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout)
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
