package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestMediaItemDataSourceReadSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/media/7" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": 7,
			"tmdbId": 550,
			"tvdbId": 12345,
			"mediaType": "movie",
			"status": 5,
			"status4k": 1
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &MediaItemDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := MediaItemDataSourceModel{ID: types.StringValue("7")}

	if err := d.refreshMediaItem(context.Background(), &data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := data.TMDBID.ValueInt64(); got != 550 {
		t.Fatalf("expected tmdb_id 550, got %d", got)
	}
	if got := data.TVDBID.ValueInt64(); got != 12345 {
		t.Fatalf("expected tvdb_id 12345, got %d", got)
	}
	if got := data.MediaType.ValueString(); got != "movie" {
		t.Fatalf("expected media_type movie, got %q", got)
	}
	if got := data.Status.ValueInt64(); got != 5 {
		t.Fatalf("expected status 5, got %d", got)
	}
	if got := data.Status4k.ValueInt64(); got != 1 {
		t.Fatalf("expected status_4k 1, got %d", got)
	}
	if data.ResponseJSON.IsNull() || data.ResponseJSON.ValueString() == "" {
		t.Fatal("expected response_json to be populated")
	}
}

func TestMediaItemDataSourceRead404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/media/999" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Media does not exist."}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &MediaItemDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := MediaItemDataSourceModel{ID: types.StringValue("999")}

	err = d.refreshMediaItem(context.Background(), &data)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") && !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected not-found diagnostic, got %v", err)
	}
}
