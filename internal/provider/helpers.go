package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const defaultRequestTimeout = 2 * time.Minute

func normalizeRequestTimeout(timeout time.Duration) time.Duration {
	if timeout <= 0 {
		return defaultRequestTimeout
	}
	return timeout
}

func normalizeServerIdentity(id, fallbackID types.String, serverID, fallbackServerID types.Int64) (types.String, types.Int64) {
	if serverID.IsNull() || serverID.IsUnknown() {
		serverID = fallbackServerID
	}

	if id.IsNull() || id.IsUnknown() || strings.TrimSpace(id.ValueString()) == "" {
		switch {
		case !serverID.IsNull() && !serverID.IsUnknown():
			id = types.StringValue(strconv.FormatInt(serverID.ValueInt64(), 10))
		default:
			id = fallbackID
		}
	}

	return id, serverID
}

func mapFromTypesMap(ctx context.Context, input types.Map) (map[string]string, diag.Diagnostics) {
	out := map[string]string{}
	if input.IsNull() || input.IsUnknown() {
		return out, nil
	}

	var headers map[string]string
	diags := input.ElementsAs(ctx, &headers, false)
	if diags.HasError() {
		return out, diags
	}
	return headers, diags
}

func mergeJSON(base map[string]any, overrideJSON string) (map[string]any, error) {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	if overrideJSON == "" {
		return out, nil
	}

	var override map[string]any
	if err := json.Unmarshal([]byte(overrideJSON), &override); err != nil {
		return nil, err
	}
	for k, v := range override {
		out[k] = v
	}
	return out, nil
}

func findByIDInJSONArray(body []byte, id int64) ([]byte, bool, error) {
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err != nil {
		return nil, false, err
	}
	for _, item := range arr {
		raw, ok := item["id"]
		if !ok {
			continue
		}
		switch v := raw.(type) {
		case float64:
			if int64(v) == id {
				b, err := json.Marshal(item)
				return b, true, err
			}
		case int64:
			if v == id {
				b, err := json.Marshal(item)
				return b, true, err
			}
		}
	}
	return nil, false, nil
}

func requireInt64ID(raw string) (int64, error) {
	var v int64
	if _, err := fmt.Sscanf(raw, "%d", &v); err != nil {
		return 0, fmt.Errorf("invalid id %q", raw)
	}
	return v, nil
}

func buildArrBaseURL(rawURL, hostname string, port int64, useSSL bool, baseURL string) (string, error) {
	if strings.TrimSpace(rawURL) != "" {
		u, err := url.Parse(rawURL)
		if err != nil {
			return "", err
		}
		if u.Scheme == "" {
			if useSSL {
				u.Scheme = "https"
			} else {
				u.Scheme = "http"
			}
		}
		return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, strings.TrimRight(u.Path, "/")), nil
	}

	if strings.TrimSpace(hostname) == "" {
		return "", fmt.Errorf("hostname is required when url is not provided")
	}

	scheme := "http"
	if useSSL {
		scheme = "https"
	}
	host := hostname
	if port > 0 {
		host = hostname + ":" + strconv.FormatInt(port, 10)
	}
	path := strings.TrimSpace(baseURL)
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return fmt.Sprintf("%s://%s%s", scheme, host, strings.TrimRight(path, "/")), nil
}

func fetchArrProfiles(ctx context.Context, rawURL, hostname string, port int64, useSSL bool, baseURL, apiKey string, timeout time.Duration) ([]map[string]any, string, error) {
	base, err := buildArrBaseURL(rawURL, hostname, port, useSSL, baseURL)
	if err != nil {
		return nil, "", err
	}
	profilesURL := strings.TrimRight(base, "/") + "/api/v3/qualityprofile"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, profilesURL, nil)
	if err != nil {
		return nil, profilesURL, err
	}
	req.Header.Set("X-Api-Key", apiKey)

	client := &http.Client{Timeout: normalizeRequestTimeout(timeout)}
	resp, err := client.Do(req)
	if err != nil {
		return nil, profilesURL, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, profilesURL, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, profilesURL, fmt.Errorf("status %d from %s", resp.StatusCode, profilesURL)
	}

	var profiles []map[string]any
	if err := json.Unmarshal(body, &profiles); err != nil {
		return nil, profilesURL, err
	}

	return profiles, profilesURL, nil
}

type arrProfileMatch struct {
	ID   int64
	Name string
	Body []byte
}

func findArrProfile(ctx context.Context, rawURL, hostname string, port int64, useSSL bool, baseURL, apiKey string, timeout time.Duration, profileID *int64, profileName *string) (*arrProfileMatch, error) {
	profiles, profilesURL, err := fetchArrProfiles(ctx, rawURL, hostname, port, useSSL, baseURL, apiKey, timeout)
	if err != nil {
		return nil, err
	}

	var trimmedName string
	if profileName != nil {
		trimmedName = strings.TrimSpace(*profileName)
	}

	for _, p := range profiles {
		rawID, ok := p["id"]
		if !ok {
			continue
		}
		var id int64
		switch v := rawID.(type) {
		case float64:
			id = int64(v)
		case int64:
			id = v
		default:
			continue
		}

		name, ok := p["name"].(string)
		if !ok || strings.TrimSpace(name) == "" {
			continue
		}

		matchesID := profileID != nil && id == *profileID
		matchesName := profileName != nil && strings.TrimSpace(name) == trimmedName
		if !matchesID && !matchesName {
			continue
		}

		body, err := json.Marshal(p)
		if err != nil {
			return nil, err
		}

		return &arrProfileMatch{
			ID:   id,
			Name: strings.TrimSpace(name),
			Body: body,
		}, nil
	}

	switch {
	case profileID != nil:
		return nil, fmt.Errorf("profile id %d not found at %s", *profileID, profilesURL)
	case profileName != nil:
		return nil, fmt.Errorf("profile %q not found at %s", trimmedName, profilesURL)
	default:
		return nil, fmt.Errorf("profile selector is required")
	}
}

// HandleAPIResponse checks the API response status and adds errors to diagnostics if not OK.
func HandleAPIResponse(ctx context.Context, resp *APIResponse, diags *diag.Diagnostics, action string) bool {
	if resp == nil {
		diags.AddError(action+" Failed", "The API returned no response.")
		return false
	}

	if StatusIsOK(resp.StatusCode) {
		return true
	}

	// Try to extract a more descriptive error from the body
	var errBody struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}

	errorMsg := string(resp.Body)
	if err := json.Unmarshal(resp.Body, &errBody); err == nil {
		if errBody.Message != "" {
			errorMsg = errBody.Message
		} else if errBody.Error != "" {
			errorMsg = errBody.Error
		}
	}

	diags.AddError(
		action+" Failed",
		fmt.Sprintf("Status %d: %s", resp.StatusCode, errorMsg),
	)

	return false
}

// ValidateArrConnectivity checks if the Arr server is reachable and the API key is valid.
func ValidateArrConnectivity(ctx context.Context, rawURL, hostname string, port int64, useSSL bool, baseURL, apiKey string, timeout time.Duration) error {
	base, err := buildArrBaseURL(rawURL, hostname, port, useSSL, baseURL)
	if err != nil {
		return err
	}
	statusURL := strings.TrimRight(base, "/") + "/api/v3/system/status"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", apiKey)

	client := &http.Client{Timeout: normalizeRequestTimeout(timeout)}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", statusURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return fmt.Errorf("invalid API key for %s", statusURL)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("status %d from %s", resp.StatusCode, statusURL)
	}

	return nil
}

