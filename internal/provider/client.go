package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type APIClient struct {
	baseURL       *url.URL
	apiKey        string
	sessionCookie string
	userAgent     string
	client        *http.Client
}

type APIResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

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

	return &APIClient{
		baseURL:   baseURL,
		apiKey:    apiKey,
		userAgent: userAgent,
		client: &http.Client{
			Transport: transport,
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
	c.apiKey = key
}

func (c *APIClient) SetSessionCookie(cookie string) {
	c.sessionCookie = cookie
}

func (c *APIClient) Request(ctx context.Context, method, path string, body string, extraHeaders map[string]string) (*APIResponse, error) {
	absURL, err := c.resolvePath(path)
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if strings.TrimSpace(body) != "" {
		reqBody = bytes.NewBufferString(body)
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(strings.TrimSpace(method)), absURL.String(), reqBody)
	if err != nil {
		return nil, err
	}

	if c.apiKey != "" {
		req.Header.Set("X-Api-Key", c.apiKey)
	} else if c.sessionCookie != "" {
		req.Header.Set("Cookie", c.sessionCookie)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if strings.TrimSpace(body) != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range extraHeaders {
		if strings.TrimSpace(k) == "" {
			continue
		}
		req.Header.Set(k, v)
	}

	res, err := c.client.Do(req)
	if err != nil {
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
