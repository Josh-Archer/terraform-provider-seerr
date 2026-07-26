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

func TestJellyfinUserImportPerformImport(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/user/import-from-jellyfin", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`[
			{"id":4,"email":"charlie@example.com","username":"Charlie","jellyfinUserId":"j-2001"},
			{"id":5,"email":"dave@example.com","username":"Dave","jellyfinUserId":"j-2002"}
		]`))
	}))
	defer mockServer.Close()

	u, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	client := NewClient(u, "test-api-key", "test-agent", false, defaultRequestTimeout)
	res := &UserImportJellyfinResource{client: client}

	data := UserImportJellyfinModel{}
	err = res.performImport(context.Background(), &data)
	require.NoError(t, err)

	assert.Equal(t, int64(2), data.ImportedCount.ValueInt64())
	assert.False(t, data.ImportedUsers.IsNull())
}
