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

func TestClientResolvePathPreservesBaseSubpath(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		path         string
		wantPath     string
		wantRawQuery string
	}{
		{
			name:     "subpath base with absolute API path",
			baseURL:  "https://example.com/seerr",
			path:     "/api/v1/settings/main",
			wantPath: "/seerr/api/v1/settings/main",
		},
		{
			name:     "root base still works",
			baseURL:  "https://example.com",
			path:     "/api/v1/settings/main",
			wantPath: "/api/v1/settings/main",
		},
		{
			name:     "base with trailing slash in path",
			baseURL:  "https://example.com/seerr/",
			path:     "/api/v1/settings/main",
			wantPath: "/seerr/api/v1/settings/main",
		},
		{
			name:     "path without leading slash",
			baseURL:  "https://example.com/seerr",
			path:     "api/v1/settings/main",
			wantPath: "/seerr/api/v1/settings/main",
		},
		{
			name:         "preserves query string on API path",
			baseURL:      "https://example.com/seerr",
			path:         "/api/v1/issue?take=1000",
			wantPath:     "/seerr/api/v1/issue",
			wantRawQuery: "take=1000",
		},
		{
			name:     "root base with trailing slash",
			baseURL:  "https://example.com/",
			path:     "/api/v1/settings/main",
			wantPath: "/api/v1/settings/main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &APIClient{baseURL: mustParseURL(t, tt.baseURL)}
			got, err := client.resolvePath(tt.path)
			if err != nil {
				t.Fatalf("resolvePath(%q) error: %v", tt.path, err)
			}
			if got.Path != tt.wantPath {
				t.Fatalf("resolvePath(%q) path = %q, want %q", tt.path, got.Path, tt.wantPath)
			}
			if got.RawQuery != tt.wantRawQuery {
				t.Fatalf("resolvePath(%q) RawQuery = %q, want %q", tt.path, got.RawQuery, tt.wantRawQuery)
			}
			if got.Scheme != client.baseURL.Scheme || got.Host != client.baseURL.Host {
				t.Fatalf("resolvePath(%q) origin = %s://%s, want %s://%s",
					tt.path, got.Scheme, got.Host, client.baseURL.Scheme, client.baseURL.Host)
			}
		})
	}
}

func TestClientRequestPreservesBaseSubpath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/seerr/api/v1/settings/main" {
			t.Fatalf("unexpected path %s, want /seerr/api/v1/settings/main", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	// Simulate reverse-proxy subpath: base URL includes /seerr.
	base, err := url.Parse(srv.URL + "/seerr")
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

func TestClientRequestAllowsSameOriginAbsoluteURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if _, err := client.Request(context.Background(), "GET", srv.URL+"/api/v1/settings/main", "", nil); err != nil {
		t.Fatalf("expected same-origin absolute URL to succeed: %v", err)
	}
}

func TestClientRequestRejectsCrossOriginAbsoluteURL(t *testing.T) {
	base, err := url.Parse("https://seerr.example.com")
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(base, "abc123", "test-agent", false, 45*time.Second)
	_, err = client.Request(context.Background(), "GET", "https://example.invalid/api/v1/settings/main", "", nil)
	if err == nil {
		t.Fatal("expected cross-origin absolute URL to be rejected")
	}
	if !strings.Contains(err.Error(), "absolute URLs must target") {
		t.Fatalf("unexpected error: %v", err)
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

func TestExtractIDFromJSON(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
		ok   bool
	}{
		{name: "string", body: `{"id":"request-42"}`, want: "request-42", ok: true},
		{name: "number", body: `{"id":42}`, want: "42", ok: true},
		{name: "blank", body: `{"id":"  "}`},
		{name: "missing", body: `{"status":"ok"}`},
		{name: "unsupported", body: `{"id":true}`},
		{name: "malformed", body: `{`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := ExtractIDFromJSON([]byte(test.body))
			if got != test.want || ok != test.ok {
				t.Fatalf("ExtractIDFromJSON(%s) = %q, %v; want %q, %v", test.body, got, ok, test.want, test.ok)
			}
		})
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
