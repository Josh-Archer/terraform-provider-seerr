package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClientRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "abc123" {
			t.Fatalf("expected X-Api-Key header")
		}
		if r.URL.Path != "/api/v1/settings/main" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	base, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	client := NewClient(base, "abc123", "test-agent", false, 45*time.Second)
	resp, err := client.Request(context.Background(), "GET", "/api/v1/settings/main", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClientSessionCookieRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Cookie") != "connect.sid=session-xyz" {
			t.Fatalf("expected connect.sid cookie in Cookie header")
		}
		if r.URL.Path != "/api/v1/settings/main" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	base, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	client := NewClient(base, "", "test-agent", false, 45*time.Second)
	client.SetSessionCookie("connect.sid=session-xyz")
	resp, err := client.Request(context.Background(), "GET", "/api/v1/settings/main", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestClientTimeoutUsesConfiguredValue(t *testing.T) {
	base, err := url.Parse("http://example.com")
	if err != nil {
		t.Fatal(err)
	}

	client := NewClient(base, "abc123", "test-agent", false, 90*time.Second)
	if got, want := client.Timeout(), 90*time.Second; got != want {
		t.Fatalf("expected timeout %s, got %s", want, got)
	}
}

func TestNormalizeRequestTimeoutFallsBackToDefault(t *testing.T) {
	if got := normalizeRequestTimeout(0); got != defaultRequestTimeout {
		t.Fatalf("expected default timeout %s, got %s", defaultRequestTimeout, got)
	}
}
