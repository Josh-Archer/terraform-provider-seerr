package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type libraryRecord struct {
	ID      any    `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func libraryIDString(id any) string {
	return fmt.Sprintf("%v", id)
}

func extractEnabledLibraryIDs(ctx context.Context, enabled types.Set) ([]string, error) {
	if enabled.IsUnknown() {
		return nil, fmt.Errorf("enabled_libraries is unknown")
	}
	if enabled.IsNull() {
		return nil, nil
	}
	var ids []string
	diags := enabled.ElementsAs(ctx, &ids, false)
	if diags.HasError() {
		return nil, fmt.Errorf("failed to extract enabled_libraries")
	}
	return ids, nil
}

// applyLibraryEnablement enables the given library IDs and disables every other
// library via PUT /{basePath}/{libraryId}. The list GET is read-only.
func applyLibraryEnablement(ctx context.Context, client *APIClient, basePath string, enabledIDs []string) ([]byte, error) {
	unlock := client.LockEndpoint(basePath)
	defer unlock()

	listBody, err := fetchLibraryList(ctx, client, basePath, false)
	if err != nil {
		return nil, err
	}

	var libs []libraryRecord
	if err := json.Unmarshal(listBody, &libs); err != nil {
		return nil, fmt.Errorf("failed to decode library list: %w", err)
	}

	enabled := make(map[string]struct{}, len(enabledIDs))
	for _, id := range enabledIDs {
		enabled[id] = struct{}{}
	}

	seen := make(map[string]struct{}, len(libs))
	for _, lib := range libs {
		seen[libraryIDString(lib.ID)] = struct{}{}
	}

	var missing []string
	for _, id := range enabledIDs {
		if _, ok := seen[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("enabled_libraries contains unknown library IDs: %s", strings.Join(missing, ", "))
	}

	for _, lib := range libs {
		id := libraryIDString(lib.ID)
		_, wantEnabled := enabled[id]
		if lib.Enabled == wantEnabled {
			continue
		}
		body, err := json.Marshal(map[string]bool{"enabled": wantEnabled})
		if err != nil {
			return nil, err
		}
		putPath := basePath + "/" + url.PathEscape(id)
		res, err := client.Request(ctx, http.MethodPut, putPath, string(body), nil)
		if err != nil {
			return nil, err
		}
		if !StatusIsOK(res.StatusCode) {
			return nil, fmt.Errorf("update library %s: status %d: %s", id, res.StatusCode, string(res.Body))
		}
	}

	return fetchLibraryList(ctx, client, basePath, false)
}

func fetchLibraryList(ctx context.Context, client *APIClient, basePath string, syncOnRead bool) ([]byte, error) {
	apiPath := basePath
	if syncOnRead {
		apiPath += "?sync=true"
	}
	res, err := client.Request(ctx, http.MethodGet, apiPath, "", nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if !StatusIsOK(res.StatusCode) {
		return nil, fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return res.Body, nil
}
