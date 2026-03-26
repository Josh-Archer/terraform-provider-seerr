package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"syscall"
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

func TestClientRequestRetriesConnectionRefusedAndEventuallySucceeds(t *testing.T) {
	var mu sync.Mutex
	attempts := 0

	client := &APIClient{
		baseURL: mustParseURL(t, "http://example.com"),
		transport: &authTransport{
			apiKey: "abc123",
		},
		client: &http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				mu.Lock()
				defer mu.Unlock()
				attempts++
				if attempts < 3 {
					return nil, &url.Error{
						Op:  req.Method,
						URL: req.URL.String(),
						Err: syscall.ECONNREFUSED,
					}
				}

				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     make(http.Header),
					Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
				}, nil
			}),
			Timeout: 2 * time.Second,
		},
	}

	resp, err := client.Request(context.Background(), "GET", "/api/v1/settings/main", "", nil)
	if err != nil {
		t.Fatalf("expected retry to succeed, got error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestClientRequestDoesNotRetryUnsafeMethods(t *testing.T) {
	attempts := 0
	client := &APIClient{
		baseURL: mustParseURL(t, "http://example.com"),
		client: &http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				return nil, &url.Error{
					Op:  req.Method,
					URL: req.URL.String(),
					Err: syscall.ECONNREFUSED,
				}
			}),
			Timeout: 2 * time.Second,
		},
	}

	_, err := client.Request(context.Background(), "POST", "/api/v1/request", `{"x":1}`, nil)
	if err == nil {
		t.Fatal("expected error for failed POST request")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt for POST, got %d", attempts)
	}
}

func TestClientRequestDoesNotRetryContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	attempts := 0
	client := &APIClient{
		baseURL: mustParseURL(t, "http://example.com"),
		client: &http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				attempts++
				return nil, context.Canceled
			}),
			Timeout: 2 * time.Second,
		},
	}

	_, err := client.Request(ctx, "GET", "/api/v1/settings/main", "", nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt after context cancellation, got %d", attempts)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse url %q: %v", raw, err)
	}

	return parsed
}
