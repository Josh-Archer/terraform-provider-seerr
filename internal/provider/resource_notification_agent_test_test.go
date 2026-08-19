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

func TestNotificationAgentTestExecute(t *testing.T) {
	testCalled := false

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/notifications/ntfy/test" && r.Method == "POST" {
			testCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"message": "Test notification sent"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockServer.Close()

	baseURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &NotificationAgentTestResource{client: client}

	ctx := context.Background()

	err = res.testNotificationAgent(ctx, "ntfy")
	require.NoError(t, err)
	assert.True(t, testCalled)
}
