package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestRequestApprovalResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewRequestApprovalResource(), map[string]resourceAttributeContract{
		"id":          computedString,
		"request_id":  requiredInt64,
		"status":      requiredString,
		"modified_by": computedInt64,
		"updated_at":  computedString,
	}, nil)
}

func TestRequestApprovalApproveEndpoint(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":42,"status":2,"updatedAt":"2026-01-01T00:00:00.000Z","modifiedBy":{"id":1}}`))
			return
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":42,"status":2,"updatedAt":"2026-01-01T00:00:00.000Z","modifiedBy":{"id":1}}`))
			return
		}
		t.Fatalf("unexpected method %s", r.Method)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &RequestApprovalResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	data := RequestApprovalModel{
		RequestID: types.Int64Value(42),
		Status:    types.StringValue("approved"),
	}

	if err := resource.readRequestDetails(context.Background(), &data); err != nil {
		t.Fatal(err)
	}

	if gotPath != "/api/v1/request/42" {
		t.Fatalf("expected GET path /api/v1/request/42, got %s", gotPath)
	}
	if got := data.Status.ValueString(); got != "approved" {
		t.Fatalf("expected status approved, got %q", got)
	}
	if got := data.ModifiedBy.ValueInt64(); got != 1 {
		t.Fatalf("expected modified_by 1, got %d", got)
	}
	if got := data.UpdatedAt.ValueString(); got != "2026-01-01T00:00:00.000Z" {
		t.Fatalf("expected updated_at timestamp, got %q", got)
	}
}

func TestRequestApprovalDeclineEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":42,"status":3,"updatedAt":"2026-01-02T00:00:00.000Z","modifiedBy":2}`))
			return
		}
		t.Fatalf("unexpected method %s", r.Method)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &RequestApprovalResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	data := RequestApprovalModel{
		RequestID: types.Int64Value(42),
		Status:    types.StringValue("decline"),
	}

	if err := resource.readRequestDetails(context.Background(), &data); err != nil {
		t.Fatal(err)
	}

	if got := data.Status.ValueString(); got != "decline" {
		t.Fatalf("expected status decline, got %q", got)
	}
	if got := data.ModifiedBy.ValueInt64(); got != 2 {
		t.Fatalf("expected modified_by 2, got %d", got)
	}
}

func TestRequestApprovalNormalizeStatus(t *testing.T) {
	if got := normalizeApprovalStatus("approved"); got != "approve" {
		t.Fatalf("expected approve, got %q", got)
	}
	if got := normalizeApprovalStatus("approve"); got != "approve" {
		t.Fatalf("expected approve, got %q", got)
	}
	if got := normalizeApprovalStatus("declined"); got != "decline" {
		t.Fatalf("expected decline, got %q", got)
	}
	if got := normalizeApprovalStatus("decline"); got != "decline" {
		t.Fatalf("expected decline, got %q", got)
	}
}

func TestRequestApprovalRead404MarksStateMissing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &RequestApprovalResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	data := RequestApprovalModel{
		ID:        types.StringValue("42"),
		RequestID: types.Int64Value(42),
	}

	if err := resource.readRequestDetails(context.Background(), &data); err != nil {
		t.Fatal(err)
	}
	if !data.ID.IsNull() {
		t.Fatalf("expected ID to be null after 404, got %q", data.ID.ValueString())
	}
}
