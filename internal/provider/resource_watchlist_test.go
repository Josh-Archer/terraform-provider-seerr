package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWatchlistResource_Schema(t *testing.T) {
	r := NewWatchlistResource()
	var resp resource.SchemaResponse
	r.Schema(t.Context(), resource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics error: %v", resp.Diagnostics)
	}

	attrs := []string{"id", "tmdb_id", "media_type", "title", "overview"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Expected '%s' attribute in schema", attr)
		}
	}
}

func TestWatchlistResource_Metadata(t *testing.T) {
	r := NewWatchlistResource()
	var metaResp resource.MetadataResponse
	r.Metadata(t.Context(), resource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_watchlist" {
		t.Errorf("Expected type name seerr_watchlist, got %s", metaResp.TypeName)
	}
}

func TestWatchlistResourceSchemaRequiresReplace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	r := NewWatchlistResource()
	var resp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &resp)
	require.False(t, resp.Diagnostics.HasError())

	nonNullState := tfsdk.State{Raw: tftypes.NewValue(tftypes.String, "state")}
	nonNullPlan := tfsdk.Plan{Raw: tftypes.NewValue(tftypes.String, "plan")}

	// tmdb_id requires replace
	tmdbAttr, ok := resp.Schema.Attributes["tmdb_id"].(schema.Int64Attribute)
	require.True(t, ok, "tmdb_id attribute not found or not Int64Attribute")
	require.NotEmpty(t, tmdbAttr.PlanModifiers, "tmdb_id missing plan modifiers")

	hasInt64RequiresReplace := false
	for _, mod := range tmdbAttr.PlanModifiers {
		if strings.Contains(mod.Description(ctx), "destroy and recreate") {
			hasInt64RequiresReplace = true
		}
	}
	assert.True(t, hasInt64RequiresReplace, "tmdb_id must have RequiresReplace plan modifier")

	int64Req := planmodifier.Int64Request{
		State:      nonNullState,
		Plan:       nonNullPlan,
		StateValue: types.Int64Value(123),
		PlanValue:  types.Int64Value(456),
	}
	int64Resp := planmodifier.Int64Response{}
	for _, mod := range tmdbAttr.PlanModifiers {
		mod.PlanModifyInt64(ctx, int64Req, &int64Resp)
	}
	assert.True(t, int64Resp.RequiresReplace, "tmdb_id modification should require replacement")

	int64ReqSame := planmodifier.Int64Request{
		State:      nonNullState,
		Plan:       nonNullPlan,
		StateValue: types.Int64Value(123),
		PlanValue:  types.Int64Value(123),
	}
	int64RespSame := planmodifier.Int64Response{}
	for _, mod := range tmdbAttr.PlanModifiers {
		mod.PlanModifyInt64(ctx, int64ReqSame, &int64RespSame)
	}
	assert.False(t, int64RespSame.RequiresReplace, "unchanged tmdb_id should not require replacement")

	// media_type requires replace
	mediaTypeAttr, ok := resp.Schema.Attributes["media_type"].(schema.StringAttribute)
	require.True(t, ok, "media_type attribute not found or not StringAttribute")
	require.NotEmpty(t, mediaTypeAttr.PlanModifiers, "media_type missing plan modifiers")

	hasStringRequiresReplace := false
	for _, mod := range mediaTypeAttr.PlanModifiers {
		if strings.Contains(mod.Description(ctx), "destroy and recreate") {
			hasStringRequiresReplace = true
		}
	}
	assert.True(t, hasStringRequiresReplace, "media_type must have RequiresReplace plan modifier")

	stringReq := planmodifier.StringRequest{
		State:      nonNullState,
		Plan:       nonNullPlan,
		StateValue: types.StringValue("movie"),
		PlanValue:  types.StringValue("tv"),
	}
	stringResp := planmodifier.StringResponse{}
	for _, mod := range mediaTypeAttr.PlanModifiers {
		mod.PlanModifyString(ctx, stringReq, &stringResp)
	}
	assert.True(t, stringResp.RequiresReplace, "media_type modification should require replacement")

	stringReqSame := planmodifier.StringRequest{
		State:      nonNullState,
		Plan:       nonNullPlan,
		StateValue: types.StringValue("movie"),
		PlanValue:  types.StringValue("movie"),
	}
	stringRespSame := planmodifier.StringResponse{}
	for _, mod := range mediaTypeAttr.PlanModifiers {
		mod.PlanModifyString(ctx, stringReqSame, &stringRespSame)
	}
	assert.False(t, stringRespSame.RequiresReplace, "unchanged media_type should not require replacement")
}

func TestWatchlistResource_Read_FindsItemOnSubsequentPage(t *testing.T) {
	var requestedPages []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/watchlist", r.URL.Path)
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			page = 1
		}
		requestedPages = append(requestedPages, page)

		w.Header().Set("Content-Type", "application/json")
		switch page {
		case 1:
			_, _ = w.Write([]byte(`{
				"page": 1,
				"totalPages": 2,
				"totalResults": 2,
				"results": [
					{"tmdbId": 100, "mediaType": "movie", "title": "Movie 100", "overview": "First page movie"}
				]
			}`))
		case 2:
			_, _ = w.Write([]byte(`{
				"page": 2,
				"totalPages": 2,
				"totalResults": 2,
				"results": [
					{"tmdbId": 200, "mediaType": "movie", "title": "Target Movie", "overview": "Found on page two"}
				]
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "REDACTED_TEST_VALUE", "test-agent", false, defaultRequestTimeout, 0, 0)
	r := &WatchlistResource{client: client}

	ctx := t.Context()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}

	initial := WatchlistResourceModel{
		ID:        types.StringValue("200:movie"),
		TMDBID:    types.Int64Value(200),
		MediaType: types.StringValue("movie"),
	}
	require.False(t, state.Set(ctx, &initial).HasError())

	readResp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &readResp)

	require.False(t, readResp.Diagnostics.HasError())
	assert.Equal(t, []int{1, 2}, requestedPages, "expected both pages to be queried")

	var updated WatchlistResourceModel
	require.False(t, readResp.State.Get(ctx, &updated).HasError())
	assert.False(t, updated.ID.IsNull(), "resource should not be removed from state")
	assert.Equal(t, "200:movie", updated.ID.ValueString())
	assert.Equal(t, "Target Movie", updated.Title.ValueString())
	assert.Equal(t, "Found on page two", updated.Overview.ValueString())
}

func TestWatchlistResource_Read_FindsItemOnFirstPageStopsPaginating(t *testing.T) {
	var requestedPages []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/watchlist", r.URL.Path)
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			page = 1
		}
		requestedPages = append(requestedPages, page)

		w.Header().Set("Content-Type", "application/json")
		switch page {
		case 1:
			_, _ = w.Write([]byte(`{
				"page": 1,
				"totalPages": 3,
				"totalResults": 3,
				"results": [
					{"tmdbId": 100, "mediaType": "tv", "name": "TV 100", "overview": "First page TV"}
				]
			}`))
		default:
			t.Fatalf("unexpected request for page %d", page)
		}
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "REDACTED_TEST_VALUE", "test-agent", false, defaultRequestTimeout, 0, 0)
	r := &WatchlistResource{client: client}

	ctx := t.Context()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}

	initial := WatchlistResourceModel{
		ID:        types.StringValue("100:tv"),
		TMDBID:    types.Int64Value(100),
		MediaType: types.StringValue("tv"),
	}
	require.False(t, state.Set(ctx, &initial).HasError())

	readResp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &readResp)

	require.False(t, readResp.Diagnostics.HasError())
	assert.Equal(t, []int{1}, requestedPages, "expected only first page to be queried")

	var updated WatchlistResourceModel
	require.False(t, readResp.State.Get(ctx, &updated).HasError())
	assert.False(t, updated.ID.IsNull())
	assert.Equal(t, "TV 100", updated.Title.ValueString())
	assert.Equal(t, "First page TV", updated.Overview.ValueString())
}

func TestWatchlistResource_Read_NotFoundAcrossPagesRemovesState(t *testing.T) {
	var requestedPages []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/watchlist", r.URL.Path)
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			page = 1
		}
		requestedPages = append(requestedPages, page)

		w.Header().Set("Content-Type", "application/json")
		switch page {
		case 1:
			_, _ = w.Write([]byte(`{
				"page": 1,
				"totalPages": 2,
				"totalResults": 2,
				"results": [
					{"tmdbId": 100, "mediaType": "movie", "title": "Movie 100"}
				]
			}`))
		case 2:
			_, _ = w.Write([]byte(`{
				"page": 2,
				"totalPages": 2,
				"totalResults": 2,
				"results": [
					{"tmdbId": 200, "mediaType": "movie", "title": "Movie 200"}
				]
			}`))
		default:
			_, _ = w.Write([]byte(`{"page": 3, "totalPages": 2, "totalResults": 2, "results": []}`))
		}
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "REDACTED_TEST_VALUE", "test-agent", false, defaultRequestTimeout, 0, 0)
	r := &WatchlistResource{client: client}

	ctx := t.Context()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}

	initial := WatchlistResourceModel{
		ID:        types.StringValue("999:movie"),
		TMDBID:    types.Int64Value(999),
		MediaType: types.StringValue("movie"),
	}
	require.False(t, state.Set(ctx, &initial).HasError())

	readResp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &readResp)

	require.False(t, readResp.Diagnostics.HasError())
	assert.Equal(t, []int{1, 2}, requestedPages, "expected both pages to be checked")

	var updated WatchlistResourceModel
	_ = readResp.State.Get(ctx, &updated)
	assert.True(t, updated.ID.IsNull(), "resource should be removed from state when not found")
}

func TestWatchlistResource_Read_404RemovesState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "REDACTED_TEST_VALUE", "test-agent", false, defaultRequestTimeout, 0, 0)
	r := &WatchlistResource{client: client}

	ctx := t.Context()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}

	initial := WatchlistResourceModel{
		ID:        types.StringValue("100:movie"),
		TMDBID:    types.Int64Value(100),
		MediaType: types.StringValue("movie"),
	}
	require.False(t, state.Set(ctx, &initial).HasError())

	readResp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &readResp)

	require.False(t, readResp.Diagnostics.HasError())
	var updated WatchlistResourceModel
	_ = readResp.State.Get(ctx, &updated)
	assert.True(t, updated.ID.IsNull(), "resource should be removed from state on 404")
}

func TestWatchlistResource_Read_ServerErrorPreservesState(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message": "internal error"}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "REDACTED_TEST_VALUE", "test-agent", false, defaultRequestTimeout, 0, 0)
	r := &WatchlistResource{client: client}

	ctx := t.Context()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}

	initial := WatchlistResourceModel{
		ID:        types.StringValue("100:movie"),
		TMDBID:    types.Int64Value(100),
		MediaType: types.StringValue("movie"),
	}
	require.False(t, state.Set(ctx, &initial).HasError())

	readResp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &readResp)

	assert.True(t, readResp.Diagnostics.HasError(), "expected error diagnostics on 500")
	var updated WatchlistResourceModel
	_ = readResp.State.Get(ctx, &updated)
	assert.False(t, updated.ID.IsNull(), "resource should be preserved in state on server error")
}

func TestWatchlistResource_Read_RawArrayFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"tmdbId": 100, "mediaType": "movie", "title": "Array Movie", "overview": "Array overview"}
		]`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "REDACTED_TEST_VALUE", "test-agent", false, defaultRequestTimeout, 0, 0)
	r := &WatchlistResource{client: client}

	ctx := t.Context()
	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	state := tfsdk.State{Schema: schemaResp.Schema}

	initial := WatchlistResourceModel{
		ID:        types.StringValue("100:movie"),
		TMDBID:    types.Int64Value(100),
		MediaType: types.StringValue("movie"),
	}
	require.False(t, state.Set(ctx, &initial).HasError())

	readResp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &readResp)

	require.False(t, readResp.Diagnostics.HasError())
	var updated WatchlistResourceModel
	require.False(t, readResp.State.Get(ctx, &updated).HasError())
	assert.Equal(t, "Array Movie", updated.Title.ValueString())
	assert.Equal(t, "Array overview", updated.Overview.ValueString())
}
