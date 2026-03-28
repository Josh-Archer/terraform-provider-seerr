package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRequestReadMarksStateMissingOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request/42" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &RequestResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := RequestModel{ID: types.StringValue("42")}

	diags := resource.readRequest(context.Background(), "42", &data)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if !data.ID.IsNull() {
		t.Fatalf("expected id to be null after 404, got %q", data.ID.ValueString())
	}
}

func TestRequestReadPopulatesComputedIDs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request/42" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": 1,
			"is4k":   false,
			"media": map[string]any{
				"id":        7,
				"mediaType": "movie",
				"tmdbId":    550,
			},
			"requestedBy": map[string]any{
				"id": 1,
			},
		})
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &RequestResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := RequestModel{ID: types.StringValue("42")}

	diags := resource.readRequest(context.Background(), "42", &data)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := data.SeerrMediaID.ValueInt64(); got != 7 {
		t.Fatalf("expected seerr_media_id 7, got %d", got)
	}
	if got := data.UserID.ValueInt64(); got != 1 {
		t.Fatalf("expected user_id 1, got %d", got)
	}
}
