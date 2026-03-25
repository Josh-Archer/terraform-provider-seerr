package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"
)

type APIClient struct {
	baseURL   *url.URL
	userAgent string
	client    *http.Client
	transport *authTransport
}

type authTransport struct {
	apiKey        string
	sessionCookie string
	userAgent     string
	next          http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.apiKey != "" {
		req.Header.Set("X-Api-Key", t.apiKey)
	} else if t.sessionCookie != "" {
		req.Header.Set("Cookie", t.sessionCookie)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", t.userAgent)
	if req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return t.next.RoundTrip(req)
}

type APIResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

const (
	maxRequestAttempts = 3
	requestRetryDelay  = 250 * time.Millisecond
)

func NewClient(baseURL *url.URL, apiKey, userAgent string, insecureSkipVerify bool, timeout time.Duration) *APIClient {
	baseTransport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		baseTransport = &http.Transport{}
	}
	transport := baseTransport.Clone()
	transport.TLSClientConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: insecureSkipVerify,
	}

	at := &authTransport{
		apiKey:    apiKey,
		userAgent: userAgent,
		next:      transport,
	}

	return &APIClient{
		baseURL:   baseURL,
		userAgent: userAgent,
		transport: at,
		client: &http.Client{
			Transport: at,
			Timeout:   normalizeRequestTimeout(timeout),
		},
	}
}

func (c *APIClient) Timeout() time.Duration {
	if c == nil || c.client == nil {
		return defaultRequestTimeout
	}
	return normalizeRequestTimeout(c.client.Timeout)
}

func (c *APIClient) SetAPIKey(key string) {
	if c.transport != nil {
		c.transport.apiKey = key
	}
}

func (c *APIClient) SetSessionCookie(cookie string) {
	if c.transport != nil {
		c.transport.sessionCookie = cookie
	}
}

func (c *APIClient) Request(ctx context.Context, method, path string, body string, extraHeaders map[string]string) (*APIResponse, error) {
	absURL, err := c.resolvePath(path)
	if err != nil {
		return nil, err
	}

	method = strings.ToUpper(strings.TrimSpace(method))
	bodyBytes := []byte(body)
	var lastErr error

	for attempt := 1; attempt <= maxRequestAttempts; attempt++ {
		if attempt > 1 {
			if err := sleepWithContext(ctx, requestRetryDelay*time.Duration(attempt-1)); err != nil {
				if lastErr != nil {
					return nil, lastErr
				}
				return nil, err
			}
		}

		var reqBody io.Reader
		if len(bodyBytes) > 0 {
			reqBody = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, absURL.String(), reqBody)
		if err != nil {
			return nil, err
		}

		for k, v := range extraHeaders {
			if strings.TrimSpace(k) == "" {
				continue
			}
			req.Header.Set(k, v)
		}

		res, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRequestAttempts && isRetryableRequestError(ctx, err) {
				continue
			}
			return nil, err
		}
		defer res.Body.Close()

		respBody, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		return &APIResponse{
			StatusCode: res.StatusCode,
			Body:       respBody,
			Headers:    res.Header,
		}, nil
	}

	return nil, lastErr
}

func (c *APIClient) resolvePath(path string) (*url.URL, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return url.Parse(path)
	}

	normalized := path
	if !strings.HasPrefix(normalized, "/") {
		normalized = "/" + normalized
	}
	return c.baseURL.Parse(normalized)
}

func StatusIsOK(code int) bool {
	return code >= 200 && code < 300
}

func ExtractIDFromJSON(body []byte) (string, bool) {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", false
	}

	raw, ok := decoded["id"]
	if !ok {
		return "", false
	}

	switch v := raw.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "", false
		}
		return v, true
	case float64:
		return fmt.Sprintf("%.0f", v), true
	default:
		return "", false
	}
}

func isRetryableRequestError(ctx context.Context, err error) bool {
	if err == nil || ctx.Err() != nil || errors.Is(err, context.Canceled) {
		return false
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	return errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.ECONNABORTED)
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
