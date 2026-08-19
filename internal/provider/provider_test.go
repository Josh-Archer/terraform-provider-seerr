package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
provider "seerr" {
  url      = "http://localhost:5055"
  api_key  = "dummy_api_key_for_testing"
}
`
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"seerr": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccClient() (*APIClient, error) {
	baseURL := os.Getenv("SEERR_URL")
	apiKey := os.Getenv("SEERR_API_KEY")

	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("SEERR_URL and SEERR_API_KEY must be set for sweepers")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return NewClient(parsed, apiKey, "terraform-provider-seerr-sweeper", false, 0, 0, 0), nil
}

func TestProviderMetadata(t *testing.T) {
	p := &SeerrProvider{version: "test"}
	resp := &provider.MetadataResponse{}
	p.Metadata(context.Background(), provider.MetadataRequest{}, resp)

	if resp.TypeName != "seerr" {
		t.Fatalf("expected type name seerr, got %s", resp.TypeName)
	}
	if resp.Version != "test" {
		t.Fatalf("expected version test, got %s", resp.Version)
	}
}

func TestResolveProviderConfigValuesUsesEnvironmentFallbacks(t *testing.T) {
	config, err := resolveProviderConfigValues(SeerrProviderModel{}, "test", func(key string) string {
		switch key {
		case "SEERR_URL":
			return "https://seerr.example.com"
		case "SEERR_API_KEY":
			return "secret"
		case "SEERR_PLEX_TOKEN":
			return "plex-secret"
		case "SEERR_REQUEST_TIMEOUT_SECONDS":
			return "45"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.BaseURL != "https://seerr.example.com" {
		t.Fatalf("expected env URL, got %q", config.BaseURL)
	}
	if config.APIKey != "secret" {
		t.Fatalf("expected env API key, got %q", config.APIKey)
	}
	if config.PlexToken != "plex-secret" {
		t.Fatalf("expected env Plex token, got %q", config.PlexToken)
	}
	if config.RequestTimeout != 45*time.Second {
		t.Fatalf("expected 45s timeout, got %s", config.RequestTimeout)
	}
}

func TestResolveProviderConfigValuesRejectsInvalidTimeout(t *testing.T) {
	_, err := resolveProviderConfigValues(SeerrProviderModel{}, "test", func(key string) string {
		if key == "SEERR_REQUEST_TIMEOUT_SECONDS" {
			return "invalid"
		}
		return ""
	})
	if err == nil {
		t.Fatal("expected invalid timeout error")
	}
}

func TestBootstrapAPIKeyFromPlexToken(t *testing.T) {
	var authCookie string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/plex":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			w.Header().Add("Set-Cookie", "connect.sid=session-123; Path=/; HttpOnly")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true}`))
		case "/api/v1/settings/main":
			authCookie = r.Header.Get("Cookie")
			if authCookie != "connect.sid=session-123" {
				t.Fatalf("expected session cookie, got %q", authCookie)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"apiKey":"fetched-key"}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	client := NewClient(baseURL, "", "test-agent", false, defaultRequestTimeout, 0, 0)
	apiKey, err := bootstrapAPIKeyFromPlexToken(context.Background(), client, "plex-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiKey != "fetched-key" {
		t.Fatalf("expected fetched-key, got %q", apiKey)
	}
}

func TestResolveProviderConfigValuesPrefersConfigValues(t *testing.T) {
	config, err := resolveProviderConfigValues(SeerrProviderModel{
		URL:            types.StringValue("https://configured.example.com"),
		APIKey:         types.StringValue("configured-key"),
		UserAgent:      types.StringValue("custom-agent"),
		RequestTimeout: types.Int64Value(90),
	}, "test", func(string) string { return "" })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.BaseURL != "https://configured.example.com" {
		t.Fatalf("expected configured URL, got %q", config.BaseURL)
	}
	if config.APIKey != "configured-key" {
		t.Fatalf("expected configured API key, got %q", config.APIKey)
	}
	if config.UserAgent != "custom-agent" {
		t.Fatalf("expected custom user agent, got %q", config.UserAgent)
	}
	if config.RequestTimeout != 90*time.Second {
		t.Fatalf("expected 90s timeout, got %s", config.RequestTimeout)
	}
}

func TestResolveProviderConfigValuesInsecureSkipVerifyEnvVar(t *testing.T) {
	config, err := resolveProviderConfigValues(SeerrProviderModel{}, "test", func(key string) string {
		if key == "SEERR_INSECURE_SKIP_VERIFY" {
			return "true"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !config.Insecure {
		t.Fatal("expected insecure=true from env var")
	}
}

func TestResolveProviderConfigValuesInsecureSkipVerifyConfigOverridesEnv(t *testing.T) {
	config, err := resolveProviderConfigValues(SeerrProviderModel{
		InsecureSkipVerify: types.BoolValue(false),
	}, "test", func(key string) string {
		if key == "SEERR_INSECURE_SKIP_VERIFY" {
			return "true"
		}
		return ""
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Insecure {
		t.Fatal("expected config false to override env true")
	}
}

func TestResolveProviderConfigValuesRetryEnvVars(t *testing.T) {
	config, err := resolveProviderConfigValues(SeerrProviderModel{}, "test", func(key string) string {
		switch key {
		case "SEERR_MAX_RETRIES":
			return "5"
		case "SEERR_RETRY_BACKOFF_SECONDS":
			return "3"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.MaxRetries != 5 {
		t.Fatalf("expected max_retries=5, got %d", config.MaxRetries)
	}
	if config.RetryBackoff != 3*time.Second {
		t.Fatalf("expected retry_backoff=3s, got %s", config.RetryBackoff)
	}
}

func TestResolveProviderConfigValuesRetryConfigOverridesEnv(t *testing.T) {
	config, err := resolveProviderConfigValues(SeerrProviderModel{
		MaxRetries:   types.Int64Value(10),
		RetryBackoff: types.Int64Value(5),
	}, "test", func(key string) string {
		switch key {
		case "SEERR_MAX_RETRIES":
			return "2"
		case "SEERR_RETRY_BACKOFF_SECONDS":
			return "1"
		default:
			return ""
		}
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.MaxRetries != 10 {
		t.Fatalf("expected max_retries=10 from config, got %d", config.MaxRetries)
	}
	if config.RetryBackoff != 5*time.Second {
		t.Fatalf("expected retry_backoff=5s from config, got %s", config.RetryBackoff)
	}
}

func TestResolveProviderConfigValuesRetryDefaults(t *testing.T) {
	config, err := resolveProviderConfigValues(SeerrProviderModel{}, "test", func(string) string { return "" })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.MaxRetries != defaultMaxRetries {
		t.Fatalf("expected default max_retries=%d, got %d", defaultMaxRetries, config.MaxRetries)
	}
	if config.RetryBackoff != defaultRetryBackoff {
		t.Fatalf("expected default retry_backoff=%s, got %s", defaultRetryBackoff, config.RetryBackoff)
	}
}

func TestResolveProviderConfigValuesRejectsInvalidRetryEnvVars(t *testing.T) {
	_, err := resolveProviderConfigValues(SeerrProviderModel{}, "test", func(key string) string {
		if key == "SEERR_MAX_RETRIES" {
			return "invalid"
		}
		return ""
	})
	if err == nil {
		t.Fatal("expected invalid max_retries error")
	}

	_, err = resolveProviderConfigValues(SeerrProviderModel{}, "test", func(key string) string {
		if key == "SEERR_RETRY_BACKOFF_SECONDS" {
			return "invalid"
		}
		return ""
	})
	if err == nil {
		t.Fatal("expected invalid retry_backoff error")
	}
}
