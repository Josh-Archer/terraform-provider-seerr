package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
)

const maxInt = math.MaxInt

// defaultPaginationPageSize is used when auto-fetching all pages from Seerr list endpoints.
// Prefer a moderate size over a large hard-coded take so servers are not overloaded.
const defaultPaginationPageSize = 100

// fetchAllPaginatedResults loads every page from a Seerr list endpoint that returns
// { "results": [...], "pageInfo": { "pages", "page", "pageSize", "results", "total" } }.
//
// basePath may include an existing query string (e.g. /api/v1/issue?filter=open); take and
// skip are set/overwritten for each page. When pageSize <= 0, defaultPaginationPageSize is used.
func fetchAllPaginatedResults(ctx context.Context, client *APIClient, basePath string, pageSize int) ([]map[string]any, error) {
	if client == nil {
		return nil, fmt.Errorf("API client is nil")
	}
	if pageSize <= 0 {
		pageSize = defaultPaginationPageSize
	}

	pathOnly, query, err := splitPathQuery(basePath)
	if err != nil {
		return nil, err
	}

	var all []map[string]any
	skip := 0
	if skipStr := query.Get("skip"); skipStr != "" {
		if initialSkip, err := strconv.Atoi(skipStr); err == nil && initialSkip > 0 {
			skip = initialSkip
		}
	}

	for {
		pageQuery := cloneURLValues(query)
		pageQuery.Set("take", strconv.Itoa(pageSize))
		pageQuery.Set("skip", strconv.Itoa(skip))

		endpoint := pathOnly
		if encoded := pageQuery.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}

		res, err := client.Request(ctx, "GET", endpoint, "", nil)
		if err != nil {
			return nil, err
		}
		if !StatusIsOK(res.StatusCode) {
			return nil, fmt.Errorf("status %d: %s", res.StatusCode, formatAPIErrorBody(res.Body))
		}

		pageResults, pageInfo, err := parsePaginatedResponse(res.Body)
		if err != nil {
			return nil, err
		}

		if len(pageResults) == 0 {
			break
		}

		all = append(all, pageResults...)

		// Stop when we know we are on the last page, when the page is short of effective pageSize,
		// or when we have accumulated the reported total.
		if shouldStopPagination(pageInfo, len(pageResults), pageSize, len(all)) {
			break
		}

		// Advance skip by the number of results returned so far.
		skip = len(all)
	}

	return all, nil
}

type pageInfo struct {
	pages    int
	page     int
	pageSize int
	results  int
	total    int
	hasPages bool
	hasPage  bool
	hasTotal bool
}

func parsePaginatedResponse(body []byte) ([]map[string]any, pageInfo, error) {
	var payload struct {
		Results  []map[string]any `json:"results"`
		PageInfo json.RawMessage  `json:"pageInfo"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		var rawArray []map[string]any
		if errArr := json.Unmarshal(body, &rawArray); errArr == nil {
			return rawArray, pageInfo{}, nil
		}
		return nil, pageInfo{}, fmt.Errorf("failed to parse paginated response: %w", err)
	}

	info := pageInfo{}
	pageInfoRaw := strings.TrimSpace(string(payload.PageInfo))
	if len(payload.PageInfo) > 0 && pageInfoRaw != "" && pageInfoRaw != "null" {
		var raw map[string]any
		if err := json.Unmarshal(payload.PageInfo, &raw); err != nil {
			return nil, pageInfo{}, fmt.Errorf("failed to parse pageInfo: %w", err)
		}
		if v, ok := safeIntFromAny(raw["pages"]); ok {
			info.pages = v
			info.hasPages = true
		}
		if v, ok := safeIntFromAny(raw["page"]); ok {
			info.page = v
			info.hasPage = true
		}
		if v, ok := safeIntFromAny(raw["pageSize"]); ok {
			info.pageSize = v
		}
		if v, ok := safeIntFromAny(raw["results"]); ok {
			info.results = v
		}
		if v, ok := safeIntFromAny(raw["total"]); ok {
			info.total = v
			info.hasTotal = true
		}
	}

	if payload.Results == nil {
		payload.Results = []map[string]any{}
	}
	return payload.Results, info, nil
}

func shouldStopPagination(info pageInfo, pageLen, pageSize, totalFetched int) bool {
	if pageLen == 0 {
		return true
	}

	// 1. If server pagination metadata includes total count:
	if info.hasTotal && info.total >= 0 {
		if totalFetched >= info.total {
			return true
		}
		// Server reported total count and totalFetched < info.total, meaning more items exist.
		// If server pageInfo explicitly says current page reaches or exceeds total pages, stop.
		if info.hasPage && info.hasPages && info.page > 0 && info.pages > 0 {
			if info.page >= info.pages {
				return true
			}
		}
		// Do not stop on short page alone when total count indicates more items remain.
		return false
	}

	// 2. If server pagination metadata has page and pages (without total):
	if info.hasPage && info.hasPages && info.page > 0 && info.pages > 0 {
		if info.page >= info.pages {
			return true
		}
		return false
	}

	// 3. Fallback heuristic (no usable pageInfo metadata available):
	// Stop if page length is less than effective page size.
	effectivePageSize := pageSize
	if info.pageSize > 0 {
		effectivePageSize = info.pageSize
	}

	return pageLen < effectivePageSize
}

// formatAPIErrorBody extracts message/error from a JSON API body when present,
// matching HandleAPIResponse diagnostics style.
func formatAPIErrorBody(body []byte) string {
	errorMsg := string(body)
	var errBody struct {
		Message string `json:"message"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &errBody); err == nil {
		if errBody.Message != "" {
			return errBody.Message
		}
		if errBody.Error != "" {
			return errBody.Error
		}
	}
	return errorMsg
}

func splitPathQuery(basePath string) (string, url.Values, error) {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" {
		return "", nil, fmt.Errorf("base path is empty")
	}

	// Use url.Parse so relative paths with queries work.
	u, err := url.Parse(basePath)
	if err != nil {
		return "", nil, fmt.Errorf("invalid base path %q: %w", basePath, err)
	}

	pathOnly := u.Path
	if pathOnly == "" {
		if i := strings.Index(basePath, "?"); i >= 0 {
			pathOnly = basePath[:i]
		} else {
			pathOnly = basePath
		}
	}

	query := url.Values{}
	for k, vs := range u.Query() {
		for _, v := range vs {
			query.Add(k, v)
		}
	}
	return pathOnly, query, nil
}

func cloneURLValues(in url.Values) url.Values {
	out := make(url.Values, len(in))
	for k, vs := range in {
		cp := make([]string, len(vs))
		copy(cp, vs)
		out[k] = cp
	}
	return out
}
