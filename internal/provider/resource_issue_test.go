package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIssueReadMarksStateMissingOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issue/11" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &IssueResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := IssueModel{ID: types.StringValue("11")}

	diags := resource.readIssue(context.Background(), "11", &data)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if !data.ID.IsNull() {
		t.Fatalf("expected id to be null after 404, got %q", data.ID.ValueString())
	}
}

func TestIssueReadPopulatesMediaID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issue/12" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":12,"issueType":4,"status":1,"media":{"id":550}}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &IssueResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := IssueModel{ID: types.StringValue("12")}

	diags := resource.readIssue(context.Background(), "12", &data)
	if diags.HasError() {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if got := data.MediaID.ValueInt64(); got != 550 {
		t.Fatalf("expected media_id 550, got %d", got)
	}
}
