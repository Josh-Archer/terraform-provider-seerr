package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

func writeUpstreamPage(w http.ResponseWriter, page, pages, total, count int, firstID int) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"page":%d,"totalPages":%d,"totalResults":%d,"results":[`, page, pages, total)
	for i := 0; i < count; i++ {
		if i > 0 {
			fmt.Fprint(w, ",")
		}
		id := firstID + i
		fmt.Fprintf(w, `{"id":%d,"ratingKey":"rk-%d","title":"Title %d","mediaType":"movie","tmdbId":%d}`, id, id, id, id)
	}
	fmt.Fprint(w, `]}`)
}

func watchlistDataSourceRequest(t *testing.T, d *WatchlistDataSource, userID *int64) datasource.ReadResponse {
	t.Helper()
	ctx := context.Background()
	var schemaResp datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schemaResp)
	typ, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok)
	vals := map[string]tftypes.Value{}
	for name, attrType := range typ.AttributeTypes {
		vals[name] = tftypes.NewValue(attrType, nil)
	}
	if userID != nil {
		vals["user_id"] = tftypes.NewValue(tftypes.Number, *userID)
	}
	req := datasource.ReadRequest{Config: tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(typ, vals)}}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: schemaResp.Schema}}
	d.Read(ctx, req, &resp)
	return resp
}

func TestWatchlistDataSourceReadUsesUserRouteAndAllPages(t *testing.T) {
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())
		require.Equal(t, "/api/v1/user/7/watchlist", r.URL.Path)
		page := 1
		_, _ = fmt.Sscan(r.URL.Query().Get("page"), &page)
		counts := []int{20, 20, 2}
		starts := []int{1, 21, 41}
		writeUpstreamPage(w, page, 3, 42, counts[page-1], starts[page-1])
	}))
	defer srv.Close()
	parsed, err := url.Parse(srv.URL)
	require.NoError(t, err)
	d := &WatchlistDataSource{client: NewClient(parsed, "test-key", "test", false, defaultRequestTimeout, 0, 0)}
	userID := int64(7)
	resp := watchlistDataSourceRequest(t, d, &userID)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	var got WatchlistDataSourceModel
	require.False(t, resp.State.Get(t.Context(), &got).HasError())
	require.Equal(t, int64(42), got.Total.ValueInt64())
	var items []WatchlistItemModel
	require.False(t, got.Watchlist.ElementsAs(t.Context(), &items, false).HasError())
	require.Len(t, items, 42)
	require.Equal(t, int64(42), items[41].TMDBID.ValueInt64())
	require.Equal(t, []string{"/api/v1/user/7/watchlist?page=1", "/api/v1/user/7/watchlist?page=2", "/api/v1/user/7/watchlist?page=3"}, paths)
}

func TestWatchlistDataSourceReadResolvesAuthenticatedUserAndEmptyList(t *testing.T) {
	for _, emptyPages := range []int{0, 1} {
		t.Run(fmt.Sprintf("totalPages_%d", emptyPages), func(t *testing.T) {
			var authReads, watchReads int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/auth/me":
					authReads++
					_, _ = w.Write([]byte(`{"id":17,"username":"test"}`))
				case "/api/v1/user/17/watchlist":
					watchReads++
					writeUpstreamPage(w, 1, emptyPages, 0, 0, 1)
				default:
					http.Error(w, "unexpected route", http.StatusNotFound)
				}
			}))
			defer srv.Close()
			parsed, err := url.Parse(srv.URL)
			require.NoError(t, err)
			d := &WatchlistDataSource{client: NewClient(parsed, "test-key", "test", false, defaultRequestTimeout, 0, 0)}
			resp := watchlistDataSourceRequest(t, d, nil)
			require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
			var got WatchlistDataSourceModel
			require.False(t, resp.State.Get(t.Context(), &got).HasError())
			require.Equal(t, int64(0), got.Total.ValueInt64())
			require.Equal(t, 1, authReads)
			require.Equal(t, 1, watchReads)
		})
	}
}

func TestWatchlistDataSourceReadDoesNotWritePartialState(t *testing.T) {
	for _, tc := range []struct {
		name         string
		secondStatus int
		secondBody   string
	}{
		{name: "later page failure", secondStatus: http.StatusBadGateway, secondBody: "upstream failure"},
		{name: "malformed later page", secondStatus: http.StatusOK, secondBody: `{"page":2,"totalPages":2,"totalResults":21}`},
		{name: "auth failure", secondStatus: http.StatusUnauthorized, secondBody: "unauthorized"},
		{name: "auth response missing id", secondStatus: http.StatusOK, secondBody: `{}`},
		{name: "auth response malformed id", secondStatus: http.StatusOK, secondBody: `{"id":"not-a-number"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/auth/me" {
					if tc.name == "auth failure" {
						w.WriteHeader(tc.secondStatus)
						_, _ = w.Write([]byte(tc.secondBody))
						return
					}
					if tc.name == "auth response missing id" || tc.name == "auth response malformed id" {
						_, _ = w.Write([]byte(tc.secondBody))
						return
					}
					_, _ = w.Write([]byte(`{"id":"17"}`))
					return
				}
				if r.URL.Query().Get("page") == "1" {
					writeUpstreamPage(w, 1, 2, 21, 20, 1)
					return
				}
				w.WriteHeader(tc.secondStatus)
				_, _ = w.Write([]byte(tc.secondBody))
			}))
			defer srv.Close()
			parsed, err := url.Parse(srv.URL)
			require.NoError(t, err)
			d := &WatchlistDataSource{client: NewClient(parsed, "test-key", "test", false, defaultRequestTimeout, 0, 0)}
			resp := watchlistDataSourceRequest(t, d, nil)
			require.True(t, resp.Diagnostics.HasError())
			require.True(t, resp.State.Raw.IsNull(), "failed reads must not publish partial state")
		})
	}
}

func TestWatchlistResourceReadUsesAuthenticatedWatchlistAndRetainsStateOnEndpointFailures(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusMethodNotAllowed} {
		t.Run(fmt.Sprintf("status_%d", status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/auth/me":
					_, _ = w.Write([]byte(`{"id":17}`))
				case "/api/v1/user/17/watchlist":
					w.WriteHeader(status)
					_, _ = w.Write([]byte("failed endpoint"))
				default:
					w.WriteHeader(status)
				}
			}))
			defer srv.Close()
			parsed, err := url.Parse(srv.URL)
			require.NoError(t, err)
			r := &WatchlistResource{client: NewClient(parsed, "test-key", "test", false, defaultRequestTimeout, 0, 0)}
			var schemaResp resource.SchemaResponse
			r.Schema(t.Context(), resource.SchemaRequest{}, &schemaResp)
			state := tfsdk.State{Schema: schemaResp.Schema}
			initial := WatchlistResourceModel{ID: types.StringValue("999:movie"), TMDBID: types.Int64Value(999), MediaType: types.StringValue("movie"), Title: types.StringValue("Keep me")}
			require.False(t, state.Set(t.Context(), &initial).HasError())
			resp := resource.ReadResponse{State: state}
			r.Read(t.Context(), resource.ReadRequest{State: state}, &resp)
			require.True(t, resp.Diagnostics.HasError())
			var got WatchlistResourceModel
			require.False(t, resp.State.Get(t.Context(), &got).HasError())
			require.Equal(t, "999:movie", got.ID.ValueString())
		})
	}
}

func TestWatchlistResourceReadRemovesOnlyAfterCompleteAbsence(t *testing.T) {
	for _, tc := range []struct {
		name         string
		secondStatus int
		secondBody   string
		wantError    bool
	}{
		{name: "complete list omits item"},
		{name: "later page failure", secondStatus: http.StatusBadGateway, secondBody: "failed", wantError: true},
		{name: "malformed later page", secondStatus: http.StatusOK, secondBody: `{"page":2,"totalPages":2,"totalResults":2}`, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var pages int
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/auth/me" {
					_, _ = w.Write([]byte(`{"id":17}`))
					return
				}
				pages++
				if r.URL.Query().Get("page") == "1" {
					writeUpstreamPage(w, 1, 2, 2, 1, 100)
					return
				}
				if tc.secondStatus != 0 {
					w.WriteHeader(tc.secondStatus)
					_, _ = w.Write([]byte(tc.secondBody))
					return
				}
				writeUpstreamPage(w, 2, 2, 2, 1, 200)
			}))
			defer srv.Close()
			parsed, err := url.Parse(srv.URL)
			require.NoError(t, err)
			r := &WatchlistResource{client: NewClient(parsed, "test-key", "test", false, defaultRequestTimeout, 0, 0)}
			var schemaResp resource.SchemaResponse
			r.Schema(t.Context(), resource.SchemaRequest{}, &schemaResp)
			state := tfsdk.State{Schema: schemaResp.Schema}
			initial := WatchlistResourceModel{ID: types.StringValue("999:movie"), TMDBID: types.Int64Value(999), MediaType: types.StringValue("movie"), Title: types.StringValue("Keep me")}
			require.False(t, state.Set(t.Context(), &initial).HasError())
			resp := resource.ReadResponse{State: state}
			r.Read(t.Context(), resource.ReadRequest{State: state}, &resp)
			require.Equal(t, tc.wantError, resp.Diagnostics.HasError())
			require.Equal(t, 2, pages)
			var got WatchlistResourceModel
			if tc.wantError {
				require.False(t, resp.State.Get(t.Context(), &got).HasError())
				require.Equal(t, "999:movie", got.ID.ValueString())
			} else {
				require.True(t, resp.State.Get(t.Context(), &got).HasError(), "complete absence should remove state")
			}
		})
	}
}

func TestWatchlistDataSourceRejectsIncompleteEmptyAndMalformedIdentity(t *testing.T) {
	for _, body := range []string{
		`{"page":1,"results":[]}`,
		`{"page":1,"totalPages":0,"totalResults":0,"results":[{"tmdbId":1,"mediaType":"movie"}]}`,
		`{"page":1,"totalPages":1,"totalResults":1,"results":[{"title":"missing identity"}]}`,
	} {
		t.Run(body, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte(body)) }))
			defer srv.Close()
			parsed, err := url.Parse(srv.URL)
			require.NoError(t, err)
			d := &WatchlistDataSource{client: NewClient(parsed, "test-key", "test", false, defaultRequestTimeout, 0, 0)}
			userID := int64(17)
			resp := watchlistDataSourceRequest(t, d, &userID)
			require.True(t, resp.Diagnostics.HasError())
			require.True(t, resp.State.Raw.IsNull())
		})
	}
}

func TestWatchlistResourceReadPreservesStateOnUnreliablePagination(t *testing.T) {
	for _, tc := range []struct{ name, first, second string }{
		{name: "missing page", first: `{"totalPages":1,"totalResults":0,"results":[]}`},
		{name: "missing page count", first: `{"page":1,"totalResults":0,"results":[]}`},
		{name: "missing total", first: `{"page":1,"totalPages":1,"results":[]}`},
		{name: "null results", first: `{"page":1,"totalPages":1,"totalResults":0,"results":null}`},
		{name: "negative total", first: `{"page":1,"totalPages":1,"totalResults":-1,"results":[]}`},
		{name: "negative pages", first: `{"page":1,"totalPages":-1,"totalResults":0,"results":[]}`},
		{name: "empty total with rows", first: `{"page":1,"totalPages":0,"totalResults":0,"results":[{"tmdbId":1,"mediaType":"movie"}]}`},
		{name: "invalid empty page count", first: `{"page":1,"totalPages":2,"totalResults":0,"results":[]}`},
		{name: "nonempty without pages", first: `{"page":1,"totalPages":0,"totalResults":1,"results":[{"tmdbId":1,"mediaType":"movie"}]}`},
		{name: "page limit", first: `{"page":1,"totalPages":501,"totalResults":10020,"results":[]}`},
		{name: "missing item identity", first: `{"page":1,"totalPages":1,"totalResults":1,"results":[{"title":"unknown item"}]}`},
		{name: "invalid item identity", first: `{"page":1,"totalPages":1,"totalResults":1,"results":[{"tmdbId":0,"mediaType":"movie"}]}`},
		{name: "invalid media type", first: `{"page":1,"totalPages":1,"totalResults":1,"results":[{"tmdbId":1,"mediaType":"episode"}]}`},
		{name: "truncated result count", first: `{"page":1,"totalPages":1,"totalResults":2,"results":[{"tmdbId":1,"mediaType":"movie"}]}`},
		{name: "invalid JSON", first: `{`},
		{name: "repeated page", second: `{"page":1,"totalPages":2,"totalResults":2,"results":[{"tmdbId":2,"mediaType":"movie"}]}`},
		{name: "total changed", second: `{"page":2,"totalPages":2,"totalResults":3,"results":[{"tmdbId":2,"mediaType":"movie"}]}`},
		{name: "page count changed", second: `{"page":2,"totalPages":3,"totalResults":2,"results":[{"tmdbId":2,"mediaType":"movie"}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				if req.URL.Path == "/api/v1/auth/me" {
					_, _ = w.Write([]byte(`{"id":17}`))
					return
				}
				if req.Method != http.MethodGet || req.URL.Path != "/api/v1/user/17/watchlist" {
					t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
					http.NotFound(w, req)
					return
				}
				if tc.second != "" {
					if req.URL.Query().Get("page") == "1" {
						writeUpstreamPage(w, 1, 2, 2, 1, 1)
						return
					}
					_, _ = w.Write([]byte(tc.second))
					return
				}
				_, _ = w.Write([]byte(tc.first))
			}))
			defer srv.Close()
			r := &WatchlistResource{client: testAPIClient(t, srv.URL)}
			var schemaResp resource.SchemaResponse
			r.Schema(t.Context(), resource.SchemaRequest{}, &schemaResp)
			state := tfsdk.State{Schema: schemaResp.Schema}
			initial := WatchlistResourceModel{ID: types.StringValue("999:movie"), TMDBID: types.Int64Value(999), MediaType: types.StringValue("movie"), Title: types.StringValue("Keep me"), Overview: types.StringValue("Preserve prior state")}
			require.False(t, state.Set(t.Context(), &initial).HasError())
			resp := resource.ReadResponse{State: state}
			r.Read(t.Context(), resource.ReadRequest{State: state}, &resp)
			require.True(t, resp.Diagnostics.HasError(), "unreliable enumeration must produce an error")
			var got WatchlistResourceModel
			require.False(t, resp.State.Get(t.Context(), &got).HasError())
			require.Equal(t, initial, got, "unreliable enumeration must preserve the complete prior state")
		})
	}
}
