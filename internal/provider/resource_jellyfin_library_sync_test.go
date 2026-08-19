package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJellyfinLibrarySyncTrigger(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/settings/jellyfin/sync", r.URL.Path)
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"running": true, "progress": 5, "total": 50}`))
			return
		}
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"running": false, "progress": 50, "total": 50}`))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &JellyfinLibrarySyncResource{client: client}

	ctx := context.Background()
	err = res.triggerScan(ctx, true)
	require.NoError(t, err)

	var data JellyfinLibrarySyncModel
	err = res.readStatus(ctx, &data)
	require.NoError(t, err)

	assert.False(t, data.Running.ValueBool())
	assert.Equal(t, float64(50), data.Progress.ValueFloat64())
	assert.Equal(t, float64(50), data.Total.ValueFloat64())
}
