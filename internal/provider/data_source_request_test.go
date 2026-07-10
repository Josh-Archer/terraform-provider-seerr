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

func TestRequestDataSourceReadSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/request/42" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": 42,
			"status": 2,
			"is4k": true,
			"media": {
				"id": 7,
				"mediaType": "movie",
				"tmdbId": 550
			},
			"requestedBy": {
				"id": 1
			}
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &RequestDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := RequestDataSourceModel{ID: types.StringValue("42")}

	if err := d.refreshRequest(context.Background(), &data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := data.Status.ValueInt64(); got != 2 {
		t.Fatalf("expected status 2, got %d", got)
	}
	if got := data.MediaID.ValueInt64(); got != 550 {
		t.Fatalf("expected media_id (tmdb) 550, got %d", got)
	}
	if got := data.SeerrMediaID.ValueInt64(); got != 7 {
		t.Fatalf("expected seerr_media_id 7, got %d", got)
	}
	if got := data.MediaType.ValueString(); got != "movie" {
		t.Fatalf("expected media_type movie, got %q", got)
	}
	if got := data.Is4K.ValueBool(); !got {
		t.Fatalf("expected is_4k true, got %v", got)
	}
	if got := data.UserID.ValueInt64(); got != 1 {
		t.Fatalf("expected user_id 1, got %d", got)
	}
	if data.ResponseJSON.IsNull() || data.ResponseJSON.ValueString() == "" {
		t.Fatal("expected response_json to be populated")
	}
}

func TestRequestDataSourceRead404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request/999" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Request does not exist."}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &RequestDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := RequestDataSourceModel{ID: types.StringValue("999")}

	err = d.refreshRequest(context.Background(), &data)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") && !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected not-found diagnostic, got %v", err)
	}
}
