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

func TestJobRunExecuteAndCancel(t *testing.T) {
	runCalled := false
	cancelCalled := false

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/jobs/plex-sync/run" && r.Method == "POST" {
			runCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message": "Job triggered"}`))
			return
		}
		if r.URL.Path == "/api/v1/settings/jobs/plex-sync/cancel" && r.Method == "POST" {
			cancelCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message": "Job cancelled"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &JobRunResource{client: client}

	ctx := context.Background()

	err = res.runJob(ctx, "plex-sync")
	require.NoError(t, err)
	assert.True(t, runCalled)

	err = res.cancelJob(ctx, "plex-sync")
	require.NoError(t, err)
	assert.True(t, cancelCalled)
}
