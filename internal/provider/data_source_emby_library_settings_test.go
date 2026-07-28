package provider

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmbyLibrarySettingsDataSource(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/emby/library" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`[
				{"id": "101", "name": "4K Movies", "enabled": true}
			]`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout)
	ds := &EmbyLibrarySettingsDataSource{client: client}
	assert.NotNil(t, ds)
}
