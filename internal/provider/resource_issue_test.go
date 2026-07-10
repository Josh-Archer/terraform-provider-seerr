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

// Create path that resolves an issue must fail when Seerr rejects the status POST with HTTP 400.
func TestApplyIssueStatusResolvedHTTP400(t *testing.T) {
	var gotPath string
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"cannot resolve issue"}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resource := &IssueResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	err = resource.applyIssueStatus(context.Background(), "11", 2)
	if err == nil {
		t.Fatal("expected error when POST /resolved returns 400, got nil")
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/api/v1/issue/11/resolved" {
		t.Fatalf("expected resolved endpoint, got %s", gotPath)
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected status code in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "cannot resolve issue") {
		t.Fatalf("expected response body in error, got %v", err)
	}
}

// Update path that changes status to resolved must fail when Seerr returns HTTP 500.
func TestApplyIssueStatusResolvedHTTP500(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"internal error"}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resource := &IssueResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	err = resource.applyIssueStatus(context.Background(), "42", 2)
	if err == nil {
		t.Fatal("expected error when POST /resolved returns 500, got nil")
	}
	if gotPath != "/api/v1/issue/42/resolved" {
		t.Fatalf("expected resolved endpoint, got %s", gotPath)
	}
	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected status code in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "internal error") {
		t.Fatalf("expected response body in error, got %v", err)
	}
}

func TestApplyIssueStatusOpenHTTPError(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`cannot reopen`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resource := &IssueResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	err = resource.applyIssueStatus(context.Background(), "7", 1)
	if err == nil {
		t.Fatal("expected error when POST /open returns 400, got nil")
	}
	if gotPath != "/api/v1/issue/7/open" {
		t.Fatalf("expected open endpoint, got %s", gotPath)
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func TestApplyIssueStatusOK(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":2}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	resource := &IssueResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	if err := resource.applyIssueStatus(context.Background(), "9", 2); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v1/issue/9/resolved" {
		t.Fatalf("expected resolved endpoint, got %s", gotPath)
	}
}

func TestIssueStatusPath(t *testing.T) {
	tests := map[int64]string{
		1: "open",
		2: "resolved",
	}
	for status, want := range tests {
		got, ok := issueStatusPath(status)
		if !ok || got != want {
			t.Fatalf("issueStatusPath(%d) = %q, %v; want %q, true", status, got, ok, want)
		}
	}
	if _, ok := issueStatusPath(99); ok {
		t.Fatal("expected unknown status to be rejected")
	}
}
