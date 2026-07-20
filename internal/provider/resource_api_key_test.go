package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestAPIKeyRegenerateFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/main/regenerate" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &APIKeyResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	data := APIKeyModel{}
	diags := resource.regenerateKey(context.Background(), &data)
	if !diags.HasError() {
		t.Fatal("expected error diagnostics, got none")
	}
}
