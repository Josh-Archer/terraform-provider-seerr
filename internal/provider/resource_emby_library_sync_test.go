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

func TestEmbyLibrarySyncTriggerAndStatus(t *testing.T) {
	triggered := false
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/emby/sync" && r.Method == "POST" {
			triggered = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"running": true}`))
			return
		}
		if r.URL.Path == "/api/v1/settings/emby/sync" && r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"running": false, "progress": 100, "total": 500}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout)
	res := &EmbyLibrarySyncResource{client: client}

	ctx := context.Background()
	err = res.triggerScan(ctx, true)
	require.NoError(t, err)
	assert.True(t, triggered)

	model := &EmbyLibrarySyncModel{}
	err = res.readStatus(ctx, model)
	require.NoError(t, err)
	assert.False(t, model.Running.ValueBool())
	assert.Equal(t, float64(100), model.Progress.ValueFloat64())
	assert.Equal(t, float64(500), model.Total.ValueFloat64())
}
