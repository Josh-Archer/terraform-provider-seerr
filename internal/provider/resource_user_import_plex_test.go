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

func TestPlexUserImportPerformImport(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/user/import-from-plex", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`[
			{"id":2,"email":"alice@example.com","username":"Alice","plexId":"1001"},
			{"id":3,"email":"bob@example.com","username":"Bob","plexId":"1002"}
		]`))
	}))
	defer mockServer.Close()

	u, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(u, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	res := &UserImportPlexResource{client: client}

	data := UserImportPlexModel{}
	err = res.performImport(context.Background(), &data)
	require.NoError(t, err)

	assert.Equal(t, int64(2), data.ImportedCount.ValueInt64())
	assert.False(t, data.ImportedUsers.IsNull())
}
