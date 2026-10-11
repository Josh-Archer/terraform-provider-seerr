package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func TestMetadataSettingsApplyUsesMetadatasEndpoint(t *testing.T) {
	seenPut := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/metadatas" {
			t.Errorf("unexpected path %s", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPut:
			seenPut = true
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Error(err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if payload["tv"] != "tvdb" || payload["anime"] != "tmdb" {
				t.Errorf("unexpected payload %#v", payload)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "tv": "tvdb", "anime": "tmdb", "tests": map[string]string{"tvdb": "ok", "tmdb": "ok"}})
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]any{"tv": "tvdb", "anime": "tmdb"})
		default:
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resource := &MetadataSettingsResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout, 0, 0),
	}
	data := MetadataSettingsModel{
		TV:    types.StringValue("tvdb"),
		Anime: types.StringValue("tmdb"),
	}
	if err := resource.applyMetadataSettings(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	if !seenPut {
		t.Fatal("expected PUT request")
	}
	if got := data.ID.ValueString(); got != "metadata" {
		t.Fatalf("expected metadata id, got %q", got)
	}
	require.Equal(t, "tvdb", data.TV.ValueString())
	require.Equal(t, "tmdb", data.Anime.ValueString())
	require.JSONEq(t, `{"tv":"tvdb","anime":"tmdb"}`, data.Response.ValueString())
}

func TestMetadataSettingsResourceReadUsesMetadatasEndpoint(t *testing.T) {
	ctx := context.Background()
	body := `{"tv":"tmdb","anime":"tvdb"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet || req.URL.Path != "/api/v1/settings/metadatas" {
			t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	r := &MetadataSettingsResource{client: testAPIClient(t, srv.URL)}
	state := tfsdk.State{Schema: metadataSettingsResourceSchema()}
	require.False(t, state.Set(ctx, &MetadataSettingsModel{ID: types.StringValue("metadata"), TV: types.StringValue("tvdb"), Anime: types.StringValue("tmdb")}).HasError())
	resp := resource.ReadResponse{State: state}
	r.Read(ctx, resource.ReadRequest{State: state}, &resp)
	require.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	require.False(t, resp.State.Raw.IsNull(), "refresh must retain existing settings")
	var got MetadataSettingsModel
	require.False(t, resp.State.Get(ctx, &got).HasError())
	require.Equal(t, "metadata", got.ID.ValueString())
	require.Equal(t, "tmdb", got.TV.ValueString())
	require.Equal(t, "tvdb", got.Anime.ValueString())
	require.JSONEq(t, body, got.Response.ValueString())
}

func TestMetadataSettingsApplyReportsProviderTestFailure(t *testing.T) {
	const body = `{"success":false,"tests":{"tvdb":"failed","tmdb":"not tested"}}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPut || req.URL.Path != "/api/v1/settings/metadatas" {
			t.Errorf("unexpected request %s %s", req.Method, req.URL.Path)
			http.NotFound(w, req)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	r := &MetadataSettingsResource{client: testAPIClient(t, srv.URL)}
	data := MetadataSettingsModel{ID: types.StringValue("metadata"), TV: types.StringValue("tvdb"), Anime: types.StringValue("tvdb")}
	before := data
	err := r.applyMetadataSettings(context.Background(), &data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "status 500")
	require.Contains(t, err.Error(), body)
	require.Equal(t, before, data, "failed provider tests must not update state")
}
