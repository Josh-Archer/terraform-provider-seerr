package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
)

func TestFetchAllPaginatedResultsAggregatesPages(t *testing.T) {
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/user" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		take := r.URL.Query().Get("take")
		skip := r.URL.Query().Get("skip")
		calls = append(calls, fmt.Sprintf("take=%s&skip=%s", take, skip))

		w.Header().Set("Content-Type", "application/json")
		switch skip {
		case "0":
			_, _ = w.Write([]byte(`{
				"results":[{"id":1,"email":"a@example.com"},{"id":2,"email":"b@example.com"}],
				"pageInfo":{"pages":2,"page":1,"pageSize":2,"results":2,"total":3}
			}`))
		case "2":
			_, _ = w.Write([]byte(`{
				"results":[{"id":3,"email":"c@example.com"}],
				"pageInfo":{"pages":2,"page":2,"pageSize":2,"results":1,"total":3}
			}`))
		default:
			_, _ = w.Write([]byte(`{
				"results":[],
				"pageInfo":{"pages":2,"page":3,"pageSize":2,"results":0,"total":3}
			}`))
		}
	}))
	defer srv.Close()

	client := testPaginationClient(t, srv.URL)

	results, err := fetchAllPaginatedResults(context.Background(), client, "/api/v1/user", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(results), 3; got != want {
		t.Fatalf("expected %d aggregated results, got %d", want, got)
	}
	if got, want := intFromAny(results[0]["id"]), 1; got != want {
		t.Fatalf("first result id: got %d, want %d", got, want)
	}
	if got, want := intFromAny(results[2]["id"]), 3; got != want {
		t.Fatalf("third result id: got %d, want %d", got, want)
	}
	if got, want := results[2]["email"], "c@example.com"; got != want {
		t.Fatalf("third result email: got %v, want %v", got, want)
	}

	// Should stop after last non-empty page (page 2 of 2), not request empty page 3.
	if got, want := len(calls), 2; got != want {
		t.Fatalf("expected %d API calls, got %d (%v)", want, got, calls)
	}
	if calls[0] != "take=2&skip=0" {
		t.Fatalf("first call query: got %q, want take=2&skip=0", calls[0])
	}
	if calls[1] != "take=2&skip=2" {
		t.Fatalf("second call query: got %q, want take=2&skip=2", calls[1])
	}
}

func TestFetchAllPaginatedResultsStopsOnEmptyResults(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		w.Header().Set("Content-Type", "application/json")
		if skip == 0 {
			_, _ = w.Write([]byte(`{
				"results":[{"id":10},{"id":11}],
				"pageInfo":{"pages":99,"page":1,"pageSize":2,"results":2,"total":100}
			}`))
			return
		}
		// Empty page even if pageInfo claims more pages.
		_, _ = w.Write([]byte(`{"results":[],"pageInfo":{"pages":99,"page":2,"pageSize":2,"results":0,"total":100}}`))
	}))
	defer srv.Close()

	client := testPaginationClient(t, srv.URL)

	results, err := fetchAllPaginatedResults(context.Background(), client, "/api/v1/request", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := len(results), 2; got != want {
		t.Fatalf("expected %d results, got %d", want, got)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 calls (one full page + empty), got %d", callCount)
	}
}

func TestFetchAllPaginatedResultsPreservesExistingQuery(t *testing.T) {
	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":1}],"pageInfo":{"pages":1,"page":1,"pageSize":10,"results":1,"total":1}}`))
	}))
	defer srv.Close()

	client := testPaginationClient(t, srv.URL)

	results, err := fetchAllPaginatedResults(context.Background(), client, "/api/v1/issue?filter=open&sort=added", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if gotQuery.Get("filter") != "open" {
		t.Fatalf("expected preserved filter=open, got %q", gotQuery.Get("filter"))
	}
	if gotQuery.Get("sort") != "added" {
		t.Fatalf("expected preserved sort=added, got %q", gotQuery.Get("sort"))
	}
	if gotQuery.Get("take") != "10" {
		t.Fatalf("expected take=10, got %q", gotQuery.Get("take"))
	}
	if gotQuery.Get("skip") != "0" {
		t.Fatalf("expected skip=0, got %q", gotQuery.Get("skip"))
	}
}

func TestFetchAllPaginatedResultsDefaultPageSize(t *testing.T) {
	var take string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		take = r.URL.Query().Get("take")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[],"pageInfo":{"pages":0,"page":1,"pageSize":100,"results":0,"total":0}}`))
	}))
	defer srv.Close()

	client := testPaginationClient(t, srv.URL)

	_, err := fetchAllPaginatedResults(context.Background(), client, "/api/v1/user", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if take != strconv.Itoa(defaultPaginationPageSize) {
		t.Fatalf("expected default take=%d, got %q", defaultPaginationPageSize, take)
	}
}

func TestFetchAllPaginatedResultsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	}))
	defer srv.Close()

	client := testPaginationClient(t, srv.URL)

	_, err := fetchAllPaginatedResults(context.Background(), client, "/api/v1/user", 100)
	if err == nil {
		t.Fatal("expected error on non-OK status")
	}
	if err.Error() != "status 500: boom" {
		t.Fatalf("expected extracted message in error, got %v", err)
	}
}

func TestParsePaginatedResponsePageInfoNullWithWhitespace(t *testing.T) {
	results, info, err := parsePaginatedResponse([]byte(`{"results":[{"id":1}],"pageInfo": null
}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if info.hasPages || info.hasPage || info.hasTotal {
		t.Fatalf("expected empty pageInfo for null, got %+v", info)
	}
}

func TestShouldStopPaginationShortPageWithMetadataContinues(t *testing.T) {
	// Server capped page size below requested take, but pageInfo says more pages remain.
	info := pageInfo{page: 1, pages: 3, hasPage: true, hasPages: true, total: 30, hasTotal: true}
	if shouldStopPagination(info, 10, 100, 10) {
		t.Fatal("expected short page not to stop when pageInfo indicates more pages")
	}
	if !shouldStopPagination(info, 10, 100, 30) {
		t.Fatal("expected stop when totalFetched reaches total")
	}
	// Without metadata, short page is a stop signal.
	if !shouldStopPagination(pageInfo{}, 10, 100, 10) {
		t.Fatal("expected short page to stop when pageInfo is absent")
	}
}

func TestFormatAPIErrorBody(t *testing.T) {
	if got := formatAPIErrorBody([]byte(`{"message":"nope"}`)); got != "nope" {
		t.Fatalf("message field: got %q", got)
	}
	if got := formatAPIErrorBody([]byte(`{"error":"bad"}`)); got != "bad" {
		t.Fatalf("error field: got %q", got)
	}
	if got := formatAPIErrorBody([]byte(`plain`)); got != "plain" {
		t.Fatalf("plain body: got %q", got)
	}
}

// TestUsersDataSourceReadAggregatesMultiplePages is an integration-style unit test that
// exercises the users collection data source against a multi-page mock API.
func TestUsersDataSourceReadAggregatesMultiplePages(t *testing.T) {
	const totalUsers = 150

	all := make([]map[string]any, 0, totalUsers)
	for i := 1; i <= totalUsers; i++ {
		all = append(all, map[string]any{
			"id":          i,
			"email":       fmt.Sprintf("user%d@example.com", i),
			"username":    fmt.Sprintf("user%d", i),
			"permissions": i % 8,
		})
	}

	var callSkips []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user" {
			http.NotFound(w, r)
			return
		}
		take, _ := strconv.Atoi(r.URL.Query().Get("take"))
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		callSkips = append(callSkips, skip)
		if take != defaultPaginationPageSize {
			t.Errorf("expected take=%d, got %d", defaultPaginationPageSize, take)
		}

		start := skip
		if start > len(all) {
			start = len(all)
		}
		end := start + take
		if end > len(all) {
			end = len(all)
		}
		pageResults := all[start:end]
		totalPages := (len(all) + take - 1) / take
		pageNum := 1
		if take > 0 {
			pageNum = (skip / take) + 1
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": pageResults,
			"pageInfo": map[string]any{
				"pages":    totalPages,
				"page":     pageNum,
				"pageSize": take,
				"results":  len(pageResults),
				"total":    len(all),
			},
		})
	}))
	defer srv.Close()

	ds := &UsersDataSource{client: testPaginationClient(t, srv.URL)}
	users, err := ds.fetchUsers(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(users), totalUsers; got != want {
		t.Fatalf("expected %d users aggregated, got %d", want, got)
	}
	if len(callSkips) < 2 {
		t.Fatalf("expected multi-page fetches, got skips %v", callSkips)
	}
	if callSkips[0] != 0 || callSkips[1] != defaultPaginationPageSize {
		t.Fatalf("unexpected skip sequence %v", callSkips)
	}
	if got := users[0].ID.ValueString(); got != "1" {
		t.Fatalf("first user id: got %q want 1", got)
	}
	if got := users[totalUsers-1].ID.ValueString(); got != "150" {
		t.Fatalf("last user id: got %q want 150", got)
	}
	if got := users[totalUsers-1].Email.ValueString(); got != "user150@example.com" {
		t.Fatalf("last user email: got %q", got)
	}
}

func TestFetchAllPaginatedResultsServerCappedPageSizeWithTotal(t *testing.T) {
	const totalItems = 120
	const serverCap = 25
	const requestedPageSize = 100

	items := make([]map[string]any, 0, totalItems)
	for i := 1; i <= totalItems; i++ {
		items = append(items, map[string]any{"id": i, "name": fmt.Sprintf("item-%d", i)})
	}

	var callSkips []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		callSkips = append(callSkips, skip)

		start := skip
		if start > len(items) {
			start = len(items)
		}
		end := start + serverCap
		if end > len(items) {
			end = len(items)
		}
		pageResults := items[start:end]

		w.Header().Set("Content-Type", "application/json")
		// Server returns total count but does not include pages/page metadata.
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": pageResults,
			"pageInfo": map[string]any{
				"pageSize": serverCap,
				"results":  len(pageResults),
				"total":    totalItems,
			},
		})
	}))
	defer srv.Close()

	client := testPaginationClient(t, srv.URL)
	results, err := fetchAllPaginatedResults(context.Background(), client, "/api/v1/item", requestedPageSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(results), totalItems; got != want {
		t.Fatalf("expected %d aggregated results, got %d", want, got)
	}

	// Expect 5 requests: skip 0 (25), skip 25 (25), skip 50 (25), skip 75 (25), skip 100 (20)
	expectedSkips := []int{0, 25, 50, 75, 100}
	if len(callSkips) != len(expectedSkips) {
		t.Fatalf("expected %d calls, got %d (%v)", len(expectedSkips), len(callSkips), callSkips)
	}
	for i, expected := range expectedSkips {
		if callSkips[i] != expected {
			t.Fatalf("call %d: expected skip=%d, got %d", i, expected, callSkips[i])
		}
	}
}

func TestFetchAllPaginatedResultsLargeCatalog(t *testing.T) {
	const totalItems = 550
	const pageSize = 100

	items := make([]map[string]any, 0, totalItems)
	for i := 1; i <= totalItems; i++ {
		items = append(items, map[string]any{"id": i, "name": fmt.Sprintf("catalog-item-%d", i)})
	}

	var callSkips []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		callSkips = append(callSkips, skip)

		start := skip
		if start > len(items) {
			start = len(items)
		}
		end := start + pageSize
		if end > len(items) {
			end = len(items)
		}
		pageResults := items[start:end]

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": pageResults,
			"pageInfo": map[string]any{
				"pages":    (totalItems + pageSize - 1) / pageSize,
				"page":     (skip / pageSize) + 1,
				"pageSize": pageSize,
				"results":  len(pageResults),
				"total":    totalItems,
			},
		})
	}))
	defer srv.Close()

	client := testPaginationClient(t, srv.URL)
	results, err := fetchAllPaginatedResults(context.Background(), client, "/api/v1/large", pageSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(results), totalItems; got != want {
		t.Fatalf("expected %d results, got %d", want, got)
	}
	if got, want := len(callSkips), 6; got != want {
		t.Fatalf("expected %d calls for 550 items, got %d (%v)", want, got, callSkips)
	}
}

func TestFetchAllPaginatedResultsRawArrayResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id": 1, "name": "a"}, {"id": 2, "name": "b"}]`))
	}))
	defer srv.Close()

	client := testPaginationClient(t, srv.URL)
	results, err := fetchAllPaginatedResults(context.Background(), client, "/api/v1/raw", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(results), 2; got != want {
		t.Fatalf("expected %d items, got %d", want, got)
	}
	if got, want := intFromAny(results[0]["id"]), 1; got != want {
		t.Fatalf("first item id: got %d, want %d", got, want)
	}
}

func TestRequestsDataSourceReadAggregatesMultiplePages(t *testing.T) {
	const totalRequests = 250
	all := make([]map[string]any, 0, totalRequests)
	for i := 1; i <= totalRequests; i++ {
		all = append(all, map[string]any{
			"id":          float64(i),
			"status":      float64(1),
			"media":       map[string]any{"id": float64(1000 + i)},
			"requestedBy": map[string]any{"id": float64(2000 + i)},
		})
	}

	var callSkips []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request" {
			http.NotFound(w, r)
			return
		}
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		callSkips = append(callSkips, skip)

		start := skip
		if start > len(all) {
			start = len(all)
		}
		end := start + defaultPaginationPageSize
		if end > len(all) {
			end = len(all)
		}
		pageResults := all[start:end]

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": pageResults,
			"pageInfo": map[string]any{
				"pages":    (totalRequests + defaultPaginationPageSize - 1) / defaultPaginationPageSize,
				"page":     (skip / defaultPaginationPageSize) + 1,
				"pageSize": defaultPaginationPageSize,
				"results":  len(pageResults),
				"total":    totalRequests,
			},
		})
	}))
	defer srv.Close()

	ds := &RequestsDataSource{client: testPaginationClient(t, srv.URL)}
	results, err := fetchAllPaginatedResults(context.Background(), ds.client, "/api/v1/request", defaultPaginationPageSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(results), totalRequests; got != want {
		t.Fatalf("expected %d requests, got %d", want, got)
	}
	if len(callSkips) < 3 {
		t.Fatalf("expected at least 3 page calls for 250 requests, got %v", callSkips)
	}
}

func TestIssuesDataSourceReadAggregatesMultiplePages(t *testing.T) {
	const totalIssues = 300
	all := make([]map[string]any, 0, totalIssues)
	for i := 1; i <= totalIssues; i++ {
		all = append(all, map[string]any{
			"id":        float64(i),
			"issueType": float64(1),
			"status":    float64(1),
			"media":     map[string]any{"id": float64(5000 + i)},
			"createdBy": map[string]any{"id": float64(6000 + i)},
		})
	}

	var callSkips []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/issue" {
			http.NotFound(w, r)
			return
		}
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		callSkips = append(callSkips, skip)

		start := skip
		if start > len(all) {
			start = len(all)
		}
		end := start + defaultPaginationPageSize
		if end > len(all) {
			end = len(all)
		}
		pageResults := all[start:end]

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": pageResults,
			"pageInfo": map[string]any{
				"pages":    (totalIssues + defaultPaginationPageSize - 1) / defaultPaginationPageSize,
				"page":     (skip / defaultPaginationPageSize) + 1,
				"pageSize": defaultPaginationPageSize,
				"results":  len(pageResults),
				"total":    totalIssues,
			},
		})
	}))
	defer srv.Close()

	ds := &IssuesDataSource{client: testPaginationClient(t, srv.URL)}
	results, err := fetchAllPaginatedResults(context.Background(), ds.client, "/api/v1/issue", defaultPaginationPageSize)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := len(results), totalIssues; got != want {
		t.Fatalf("expected %d issues, got %d", want, got)
	}
	if len(callSkips) < 3 {
		t.Fatalf("expected at least 3 page calls for 300 issues, got %v", callSkips)
	}
}

func testPaginationClient(t *testing.T, rawURL string) *APIClient {
	t.Helper()
	baseURL, err := url.Parse(rawURL)
	if err != nil {
		t.Fatal(err)
	}
	return NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout)
}

func intFromAny(v any) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	case json.Number:
		i, _ := n.Int64()
		return int(i)
	default:
		return -1
	}
}
