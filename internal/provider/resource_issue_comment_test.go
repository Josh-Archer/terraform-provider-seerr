package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIssueCommentResourceSchema(t *testing.T) {
	t.Parallel()
	assertResourceSchemaContract(t, NewIssueCommentResource(), map[string]resourceAttributeContract{
		"id":         computedString,
		"issue_id":   requiredInt64,
		"message":    requiredString,
		"user_id":    computedInt64,
		"created_at": computedString,
		"updated_at": computedString,
	}, nil)
}

func TestIssueCommentReadPopulatesFields(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issueComment/99" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": 99,
			"message": "Looks like audio is desynced",
			"createdAt": "2026-01-01T00:00:00.000Z",
			"updatedAt": "2026-01-02T00:00:00.000Z",
			"user": {
				"id": 4,
				"username": "admin"
			},
			"issue": {
				"id": 12
			}
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &IssueCommentResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	data := IssueCommentModel{
		ID:      types.StringValue("99"),
		IssueID: types.Int64Value(12),
	}

	if err := resource.readCommentDetails(context.Background(), &data); err != nil {
		t.Fatal(err)
	}

	if got := data.ID.ValueString(); got != "99" {
		t.Fatalf("expected ID 99, got %q", got)
	}
	if got := data.Message.ValueString(); got != "Looks like audio is desynced" {
		t.Fatalf("expected message, got %q", got)
	}
	if got := data.UserID.ValueInt64(); got != 4 {
		t.Fatalf("expected user_id 4, got %d", got)
	}
	if got := data.CreatedAt.ValueString(); got != "2026-01-01T00:00:00.000Z" {
		t.Fatalf("expected createdAt timestamp, got %q", got)
	}
	if got := data.UpdatedAt.ValueString(); got != "2026-01-02T00:00:00.000Z" {
		t.Fatalf("expected updatedAt timestamp, got %q", got)
	}
}

func TestIssueCommentReadMarksMissingOn404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &IssueCommentResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}

	data := IssueCommentModel{
		ID: types.StringValue("99"),
	}

	if err := resource.readCommentDetails(context.Background(), &data); err != nil {
		t.Fatal(err)
	}

	if !data.ID.IsNull() {
		t.Fatalf("expected ID to be null after 404, got %q", data.ID.ValueString())
	}
}
