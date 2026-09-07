package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

const libraryWriteProbeID = "__tf_seerr_probe__"

type libraryWriteMode int

const (
	libraryWriteUnknown libraryWriteMode = iota
	libraryWritePut
	libraryWriteGetEnable
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

func libraryParentPath(basePath string) string {
	return strings.TrimSuffix(basePath, "/library")
}

func decodeLibraryRecords(body []byte) ([]libraryRecord, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > 0 && trimmed[0] == '[' {
		var libs []libraryRecord
		if err := json.Unmarshal(trimmed, &libs); err != nil {
			return nil, fmt.Errorf("failed to decode library list: %w", err)
		}
		return libs, nil
	}
	var wrap struct {
		Libraries []libraryRecord `json:"libraries"`
	}
	if err := json.Unmarshal(trimmed, &wrap); err != nil {
		return nil, fmt.Errorf("failed to decode library list: %w", err)
	}
	return wrap.Libraries, nil
}

func encodeLibraryRecords(libs []libraryRecord) ([]byte, error) {
	return json.Marshal(libs)
}

func putRouteExists(status int, body []byte) bool {
	switch status {
	case http.StatusOK, http.StatusNoContent, http.StatusBadRequest:
		return true
	case http.StatusNotFound:
		return bytes.Contains(body, []byte("Library does not exist"))
	default:
		return false
	}
}

func detectLibraryWriteMode(ctx context.Context, client *APIClient, basePath string) (libraryWriteMode, error) {
	if client == nil {
		return libraryWriteUnknown, fmt.Errorf("missing API client")
	}
	if mode, ok := client.libraryWriteMode(basePath); ok {
		return mode, nil
	}
	res, err := client.Request(ctx, http.MethodPut, basePath+"/"+url.PathEscape(libraryWriteProbeID), `{"enabled":false}`, nil)
	if err != nil {
		return libraryWriteUnknown, err
	}
	mode := libraryWriteGetEnable
	if putRouteExists(res.StatusCode, res.Body) {
		mode = libraryWritePut
	}
	client.setLibraryWriteMode(basePath, mode)
	return mode, nil
}

func applyLibraryEnablement(ctx context.Context, client *APIClient, basePath string, enabledIDs []string) ([]byte, error) {
	unlock := client.LockEndpoint(basePath)
	defer unlock()

	mode, err := detectLibraryWriteMode(ctx, client, basePath)
	if err != nil {
		return nil, err
	}
	if mode == libraryWriteGetEnable {
		return writeLibrariesViaGetEnable(ctx, client, basePath, enabledIDs, false)
	}
	return writeLibrariesViaPut(ctx, client, basePath, enabledIDs)
}

func writeLibrariesViaPut(ctx context.Context, client *APIClient, basePath string, enabledIDs []string) ([]byte, error) {
	listBody, err := fetchLibraryList(ctx, client, basePath, false, nil)
	if err != nil {
		return nil, err
	}
	libs, err := decodeLibraryRecords(listBody)
	if err != nil {
		return nil, err
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

	return fetchLibraryList(ctx, client, basePath, false, enabledIDs)
}

func writeLibrariesViaGetEnable(ctx context.Context, client *APIClient, basePath string, enabledIDs []string, syncOnRead bool) ([]byte, error) {
	q := url.Values{}
	if syncOnRead {
		q.Set("sync", "true")
	}
	q.Set("enable", strings.Join(enabledIDs, ","))
	res, err := client.Request(ctx, http.MethodGet, basePath+"?"+q.Encode(), "", nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if !StatusIsOK(res.StatusCode) {
		return nil, fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	libs, err := decodeLibraryRecords(res.Body)
	if err != nil {
		return nil, err
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
	return encodeLibraryRecords(libs)
}

func fetchLibraryList(ctx context.Context, client *APIClient, basePath string, syncOnRead bool, preserveEnabled []string) ([]byte, error) {
	mode, err := detectLibraryWriteMode(ctx, client, basePath)
	if err != nil {
		return nil, err
	}
	if mode == libraryWriteGetEnable {
		if syncOnRead {
			preserve := preserveEnabled
			if len(preserve) == 0 {
				current, err := fetchLibrariesFromParent(ctx, client, basePath)
				if err != nil {
					return nil, err
				}
				for _, lib := range current {
					if lib.Enabled {
						preserve = append(preserve, libraryIDString(lib.ID))
					}
				}
			}
			return writeLibrariesViaGetEnable(ctx, client, basePath, preserve, true)
		}
		libs, err := fetchLibrariesFromParent(ctx, client, basePath)
		if err != nil {
			return nil, err
		}
		return encodeLibraryRecords(libs)
	}

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
	libs, err := decodeLibraryRecords(res.Body)
	if err != nil {
		return nil, err
	}
	return encodeLibraryRecords(libs)
}

func fetchLibrariesFromParent(ctx context.Context, client *APIClient, basePath string) ([]libraryRecord, error) {
	res, err := client.Request(ctx, http.MethodGet, libraryParentPath(basePath), "", nil)
	if err != nil {
		return nil, err
	}
	if res.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if !StatusIsOK(res.StatusCode) {
		return nil, fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	return decodeLibraryRecords(res.Body)
}
