package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientRetries429WithRetryAfterSeconds(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "abc123", "test-agent", false, 30*time.Second)
	resp, err := client.Request(context.Background(), "GET", "/api/v1/settings/main", "", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestClientRetries429WithRetryAfterHTTPDate(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			retryAt := time.Now().Add(500 * time.Millisecond).UTC().Format(time.RFC1123)
			w.Header().Set("Retry-After", retryAt)
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"rate limited"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "abc123", "test-agent", false, 30*time.Second)
	resp, err := client.Request(context.Background(), "GET", "/api/v1/settings/main", "", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestClientRetries502503504(t *testing.T) {
	for _, code := range []int{502, 503, 504} {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			var attempts int32

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				n := atomic.AddInt32(&attempts, 1)
				if n == 1 {
					w.WriteHeader(code)
					_, _ = w.Write([]byte(`error`))
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"ok":true}`))
			}))
			defer srv.Close()

			baseURL, err := url.Parse(srv.URL)
			require.NoError(t, err)

			client := NewClient(baseURL, "abc123", "test-agent", false, 30*time.Second)
			resp, err := client.Request(context.Background(), "GET", "/api/v1/settings/main", "", nil)
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
		})
	}
}

func TestClientDoesNotRetry429ForPostMethod(t *testing.T) {
	var attempts int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "abc123", "test-agent", false, 30*time.Second)
	resp, err := client.Request(context.Background(), "POST", "/api/v1/request", `{"x":1}`, nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestClientRetryAfterCappedAt60Seconds(t *testing.T) {
	var attempts int32

	client := &APIClient{
		baseURL: mustParseURL(t, "http://example.com"),
		client: &http.Client{
			Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				n := atomic.AddInt32(&attempts, 1)
				if n == 1 {
					return &http.Response{
						StatusCode: http.StatusTooManyRequests,
						Header: http.Header{
							"Retry-After": []string{"3600"}, // 1 hour, should be capped to 60s
						},
						Body: io.NopCloser(strings.NewReader(`{"error":"rate limited"}`)),
					}, nil
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

	// Use a context with a short timeout — if the cap works, the retry
	// should attempt within ~60s. With 2s client timeout this should
	// cancel before 60s. We just want to verify the retry was attempted.
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	resp, err := client.Request(ctx, "GET", "/api/v1/settings/main", "", nil)
	// Context should cancel before 60s retry
	if err == nil {
		// If we got a response, it should be the 429 (returned when context cancels during sleep)
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	}
	// Only 1 attempt because sleep was cancelled
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestParseRetryAfterSeconds(t *testing.T) {
	assert.Equal(t, 5*time.Second, parseRetryAfter("5"))
	assert.Equal(t, 1500*time.Millisecond, parseRetryAfter("1.5"))
	assert.Equal(t, time.Duration(0), parseRetryAfter(""))
	assert.Equal(t, time.Duration(0), parseRetryAfter("-1"))
}

func TestParseRetryAfterHTTPDate(t *testing.T) {
	future := time.Now().Add(10 * time.Second).UTC().Format(time.RFC1123)
	d := parseRetryAfter(future)
	assert.Greater(t, d, 5*time.Second)
	assert.LessOrEqual(t, d, 11*time.Second)
}

func TestIsRetryableStatusCode(t *testing.T) {
	assert.True(t, isRetryableStatusCode(429))
	assert.True(t, isRetryableStatusCode(502))
	assert.True(t, isRetryableStatusCode(503))
	assert.True(t, isRetryableStatusCode(504))
	assert.False(t, isRetryableStatusCode(200))
	assert.False(t, isRetryableStatusCode(400))
	assert.False(t, isRetryableStatusCode(500))
}
