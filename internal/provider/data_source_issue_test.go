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

func TestIssueDataSourceReadSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/issue/12" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": 12,
			"issueType": 4,
			"status": 1,
			"media": {"id": 550},
			"createdBy": {"id": 3}
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &IssueDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := IssueDataSourceModel{ID: types.StringValue("12")}

	if err := d.refreshIssue(context.Background(), &data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := data.IssueType.ValueInt64(); got != 4 {
		t.Fatalf("expected issue_type 4, got %d", got)
	}
	if got := data.Status.ValueInt64(); got != 1 {
		t.Fatalf("expected status 1, got %d", got)
	}
	if got := data.MediaID.ValueInt64(); got != 550 {
		t.Fatalf("expected media_id 550, got %d", got)
	}
	if got := data.CreatedByID.ValueInt64(); got != 3 {
		t.Fatalf("expected created_by_id 3, got %d", got)
	}
	if data.ResponseJSON.IsNull() || data.ResponseJSON.ValueString() == "" {
		t.Fatal("expected response_json to be populated")
	}
}

func TestIssueDataSourceRead404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issue/999" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Issue does not exist."}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &IssueDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	data := IssueDataSourceModel{ID: types.StringValue("999")}

	err = d.refreshIssue(context.Background(), &data)
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "not found") && !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected not-found diagnostic, got %v", err)
	}
}
