package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRequestRetryPostsRetryEndpoint(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":42,"status":2}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resource := &RequestRetryResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := RequestRetryModel{RequestID: types.Int64Value(42)}
	if err := resource.retryRequest(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v1/request/42/retry" {
		t.Fatalf("expected retry endpoint, got %s", gotPath)
	}
	if got := data.Status.ValueInt64(); got != 2 {
		t.Fatalf("expected status 2, got %d", got)
	}
}
