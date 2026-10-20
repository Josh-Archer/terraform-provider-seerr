package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestBlocklistResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewBlocklistResource(), map[string]resourceAttributeContract{
		"id":               computedString,
		"tmdb_id":          requiredInt64,
		"media_type":       requiredString,
		"title":            optionalComputedString,
		"user_id":          optionalComputedInt64,
		"blocklisted_tags": optionalComputedString,
		"created_at":       computedString,
	}, nil)
}

func TestBlocklistResourceSchemaRequiresReplace(t *testing.T) {
	t.Parallel()
	r := NewBlocklistResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())

	for _, attrName := range []string{"title", "blocklisted_tags", "tmdb_id", "media_type", "user_id"} {
		attr, ok := resp.Schema.Attributes[attrName]
		require.True(t, ok, "attribute %s not found", attrName)
		switch a := attr.(type) {
		case rschema.StringAttribute:
			require.NotEmpty(t, a.PlanModifiers, "string attribute %s missing plan modifiers", attrName)
		case rschema.Int64Attribute:
			require.NotEmpty(t, a.PlanModifiers, "int64 attribute %s missing plan modifiers", attrName)
		}
	}
}

func newBlocklistMock(t *testing.T, status int, body string, calls *int32) *BlocklistResource {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)
		if r.URL.Query().Get("mediaType") == "" {
			// Upstream Seerr rejects blocklist lookups without mediaType.
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	return &BlocklistResource{client: NewClient(baseURL, "k", "ua", false, defaultRequestTimeout, 0, 0)}
}

// Upstream Seerr does not persist the `user` relation from the POST body, so
// GET /blocklist/:id returns `user: null`.
func TestBlocklistRefreshKeepsConfiguredUserIDWhenAPIReturnsNull(t *testing.T) {
	t.Parallel()
	var calls int32
	r := newBlocklistMock(t, 200, `{"tmdbId":438631,"mediaType":"movie","title":"Dune","user":null,"blocklistedTags":null,"createdAt":"2025-01-01"}`, &calls)

	data := BlocklistModel{TMDBID: types.Int64Value(438631), MediaType: types.StringValue("movie"), UserID: types.Int64Value(5)}
	require.NoError(t, r.refreshBlocklist(context.Background(), &data))
	require.Equal(t, types.Int64Value(5), data.UserID)

	// After import there is no prior value: it stays null rather than unknown.
	imported := BlocklistModel{TMDBID: types.Int64Value(438631), MediaType: types.StringValue("movie"), UserID: types.Int64Null()}
	require.NoError(t, r.refreshBlocklist(context.Background(), &imported))
	require.True(t, imported.UserID.IsNull())
}

func TestBlocklistRefreshUsesAPIUserWhenPresent(t *testing.T) {
	t.Parallel()
	var calls int32
	r := newBlocklistMock(t, 200, `{"tmdbId":1,"mediaType":"tv","user":{"id":9}}`, &calls)
	data := BlocklistModel{TMDBID: types.Int64Value(1), MediaType: types.StringValue("tv"), UserID: types.Int64Value(5)}
	require.NoError(t, r.refreshBlocklist(context.Background(), &data))
	require.Equal(t, types.Int64Value(9), data.UserID)
}

func TestBlocklistRefreshDoesNotRetryWithoutMediaType(t *testing.T) {
	t.Parallel()
	var calls int32
	r := newBlocklistMock(t, 400, `{"message":"bad"}`, &calls)
	data := BlocklistModel{TMDBID: types.Int64Value(1), MediaType: types.StringValue("movie")}
	require.Error(t, r.refreshBlocklist(context.Background(), &data))
	require.EqualValues(t, 1, atomic.LoadInt32(&calls))
}
