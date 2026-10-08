package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestClientDoesNotForwardCredentialsAcrossRedirectOrigins(t *testing.T) {
	var targetRequests int
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetRequests++
		if r.Header.Get("X-Api-Key") != "" || r.Header.Get("Cookie") != "" {
			t.Errorf("credentials reached redirect target: api key present=%t, cookie present=%t", r.Header.Get("X-Api-Key") != "", r.Header.Get("Cookie") != "")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/collect", http.StatusFound)
	}))
	defer source.Close()

	baseURL, err := url.Parse(source.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, "secret-api-key", "test-agent", false, 5*time.Second, 0, 0)
	_, err = client.Request(context.Background(), http.MethodGet, "/redirect", "", nil)
	if err == nil {
		t.Fatal("expected cross-origin redirect to be rejected")
	}
	if targetRequests != 0 {
		t.Fatalf("expected redirect target to receive no request, got %d", targetRequests)
	}
}

func TestClientDoesNotForwardSessionCookieAcrossRedirectOrigins(t *testing.T) {
	var targetRequests int
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetRequests++
		if got := r.Header.Get("Cookie"); got != "" {
			t.Errorf("session cookie reached redirect target: %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/collect", http.StatusFound)
	}))
	defer source.Close()

	baseURL, err := url.Parse(source.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, "", "test-agent", false, 5*time.Second, 0, 0)
	client.SetSessionCookie("session=private")
	_, err = client.Request(context.Background(), http.MethodGet, "/redirect", "", nil)
	if err == nil {
		t.Fatal("expected cross-origin redirect to be rejected")
	}
	if targetRequests != 0 {
		t.Fatalf("expected redirect target to receive no request, got %d", targetRequests)
	}
}

func TestClientRejectsHTTPSDowngradeRedirect(t *testing.T) {
	var targetRequests int
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		targetRequests++
		w.WriteHeader(http.StatusOK)
	}))
	defer target.Close()

	source := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/collect", http.StatusFound)
	}))
	defer source.Close()

	baseURL, err := url.Parse(source.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, "secret-api-key", "test-agent", true, 5*time.Second, 0, 0)
	_, err = client.Request(context.Background(), http.MethodGet, "/redirect", "", nil)
	if err == nil {
		t.Fatal("expected HTTPS-to-HTTP redirect to be rejected")
	}
	if targetRequests != 0 {
		t.Fatalf("expected downgrade target to receive no request, got %d", targetRequests)
	}
}

func TestClientFollowsSameOriginRedirect(t *testing.T) {
	var redirectedRequestReceived bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.Redirect(w, r, "/finish", http.StatusFound)
			return
		}
		if r.URL.Path != "/finish" {
			http.NotFound(w, r)
			return
		}
		redirectedRequestReceived = true
		if got := r.Header.Get("X-Api-Key"); got != "secret-api-key" {
			t.Errorf("same-origin redirect API key: got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, "secret-api-key", "test-agent", false, 5*time.Second, 0, 0)
	response, err := client.Request(context.Background(), http.MethodGet, "/start", "", nil)
	if err != nil {
		t.Fatalf("same-origin redirect failed: %v", err)
	}
	if response.StatusCode != http.StatusOK || !redirectedRequestReceived {
		t.Fatalf("expected redirected 200 response; status=%d received=%t", response.StatusCode, redirectedRequestReceived)
	}
}

func TestClientPreservesDefaultRedirectLimit(t *testing.T) {
	var requestCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		http.Redirect(w, r, "/loop", http.StatusFound)
	}))
	defer server.Close()

	baseURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := NewClient(baseURL, "secret-api-key", "test-agent", false, 5*time.Second, 0, 0)
	_, err = client.Request(context.Background(), http.MethodGet, "/loop", "", nil)
	if err == nil {
		t.Fatal("expected redirect loop to fail after the standard redirect limit")
	}
	if requestCount != 10 {
		t.Fatalf("expected the standard limit of 10 requests, got %d", requestCount)
	}
}
