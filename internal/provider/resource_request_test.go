package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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

func TestRequestApplyStatusPostsWorkflowEndpoint(t *testing.T) {
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
	resource := &RequestResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	if err := resource.applyRequestStatus(context.Background(), "42", types.Int64Value(2)); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v1/request/42/approve" {
		t.Fatalf("expected approve endpoint, got %s", gotPath)
	}
}

func TestRequestStatusPath(t *testing.T) {
	tests := map[int64]string{
		1: "pending",
		2: "approve",
		3: "decline",
	}
	for status, want := range tests {
		got, ok := requestStatusPath(status)
		if !ok || got != want {
			t.Fatalf("requestStatusPath(%d) = %q, %v; want %q, true", status, got, ok, want)
		}
	}
	if _, ok := requestStatusPath(99); ok {
		t.Fatal("expected unknown status to be rejected")
	}
}

func TestWaitForRequestStatusSuccess(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/request/42" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		calls++
		status := 1
		if calls >= 2 {
			status = 2
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": status})
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout)

	sleepCalls := 0
	err = waitForRequestStatus(
		context.Background(),
		client,
		"42",
		2,
		5*time.Second,
		time.Second,
		func(ctx context.Context, d time.Duration) error {
			sleepCalls++
			return nil
		},
	)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if calls < 2 {
		t.Fatalf("expected at least 2 GETs, got %d", calls)
	}
	if sleepCalls < 1 {
		t.Fatalf("expected at least one sleep between polls, got %d", sleepCalls)
	}
}

func TestWaitForRequestStatusTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": 1})
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout)

	// No-op sleep: deadline uses wall clock with a short timeout so the test stays fast.
	err = waitForRequestStatus(
		context.Background(),
		client,
		"42",
		2,
		50*time.Millisecond,
		time.Millisecond,
		func(ctx context.Context, d time.Duration) error {
			return nil
		},
	)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "timed out") {
		t.Fatalf("expected timeout-related error, got %v", err)
	}
}

func TestWaitForRequestStatusContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": 1})
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel during the first sleep so the helper must honor ctx.Done().
	err = waitForRequestStatus(
		ctx,
		client,
		"42",
		2,
		5*time.Second,
		time.Second,
		func(ctx context.Context, d time.Duration) error {
			cancel()
			return ctx.Err()
		},
	)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected context cancellation error, got %v", err)
	}
}
