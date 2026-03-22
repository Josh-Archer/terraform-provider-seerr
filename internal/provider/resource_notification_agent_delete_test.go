package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNotificationDeleteConvergedWhenDisabled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/settings/notifications/pushover" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"enabled":false,"types":0,"options":{}}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := &NotificationClientResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
		agent:  "pushover",
	}

	if !r.notificationDeleteConverged(context.Background()) {
		t.Fatal("expected disabled notification agent to be treated as converged")
	}
}

func TestNotificationDeleteIgnoresTimeoutWhenAgentAlreadyDisabled(t *testing.T) {
	var requestCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if r.URL.Path != "/api/v1/settings/notifications/pushover" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		switch r.Method {
		case http.MethodPost:
			time.Sleep(100 * time.Millisecond)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"enabled":false,"types":0,"options":{}}`))
		default:
			t.Fatalf("unexpected method %s", r.Method)
		}
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	r := &NotificationClientResource{
		client: NewClient(baseURL, "abc123", "test-agent", false, 25*time.Millisecond),
		agent:  "pushover",
	}

	var resp resource.DeleteResponse
	r.Delete(context.Background(), resource.DeleteRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("expected delete to succeed after converged timeout, got %v", resp.Diagnostics)
	}
	if requestCount < 2 {
		t.Fatalf("expected timeout recovery to issue a follow-up GET, got %d requests", requestCount)
	}
}
