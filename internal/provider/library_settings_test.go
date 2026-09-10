package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLibrary struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type libraryAPICall struct {
	Method string
	Path   string
	Query  url.Values
	Body   string
}

type librarySettingsServer struct {
	*httptest.Server
	mu    sync.Mutex
	libs  []mockLibrary
	calls []libraryAPICall
}

func newLibrarySettingsServer(t *testing.T, basePath string, libs []mockLibrary) *librarySettingsServer {
	t.Helper()
	srv := &librarySettingsServer{libs: append([]mockLibrary(nil), libs...)}
	srv.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		srv.mu.Lock()
		defer srv.mu.Unlock()
		srv.calls = append(srv.calls, libraryAPICall{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.Query(),
			Body:   string(body),
		})

		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == basePath:
			// Real Seerr ignores ?enable=; mutating here would hide the original bug.
			b, err := json.Marshal(srv.libs)
			if err != nil {
				t.Errorf("marshal libraries: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, basePath+"/"):
			id := strings.TrimPrefix(r.URL.Path, basePath+"/")
			var payload struct {
				Enabled bool `json:"enabled"`
			}
			if err := json.Unmarshal(body, &payload); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			for i := range srv.libs {
				if srv.libs[i].ID == id {
					srv.libs[i].Enabled = payload.Enabled
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(srv.libs[i])
					return
				}
			}
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"Library does not exist."}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func (s *librarySettingsServer) snapshot() []mockLibrary {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]mockLibrary, len(s.libs))
	copy(out, s.libs)
	return out
}

func (s *librarySettingsServer) recordedCalls() []libraryAPICall {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]libraryAPICall, len(s.calls))
	copy(out, s.calls)
	return out
}

func testAPIClient(t *testing.T, rawURL string) *APIClient {
	t.Helper()
	baseURL, err := url.Parse(rawURL)
	require.NoError(t, err)
	return NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
}

func TestApplyLibraryEnablementPersistsAcrossRead(t *testing.T) {
	const basePath = "/api/v1/settings/jellyfin/library"
	srv := newLibrarySettingsServer(t, basePath, []mockLibrary{
		{ID: "jf1", Name: "Movies", Enabled: false},
		{ID: "jf2", Name: "TV", Enabled: false},
	})
	client := testAPIClient(t, srv.URL)
	ctx := context.Background()

	body, err := applyLibraryEnablement(ctx, client, basePath, []string{"jf1"})
	require.NoError(t, err)

	var afterUpdate []mockLibrary
	require.NoError(t, json.Unmarshal(body, &afterUpdate))
	assert.Equal(t, []mockLibrary{
		{ID: "jf1", Name: "Movies", Enabled: true},
		{ID: "jf2", Name: "TV", Enabled: false},
	}, afterUpdate)

	readBody, err := fetchLibraryList(ctx, client, basePath, false, []string{"jf1"})
	require.NoError(t, err)
	var afterRead []mockLibrary
	require.NoError(t, json.Unmarshal(readBody, &afterRead))
	assert.Equal(t, afterUpdate, afterRead)
	assert.Equal(t, afterUpdate, srv.snapshot())

	for _, call := range srv.recordedCalls() {
		assert.Empty(t, call.Query.Get("enable"), "develop GET must not write via enable query: %+v", call)
	}

	var putCount int
	for _, call := range srv.recordedCalls() {
		if call.Method == http.MethodPut && call.Path == basePath+"/jf1" {
			putCount++
			assert.JSONEq(t, `{"enabled":true}`, call.Body)
		}
	}
	assert.Equal(t, 1, putCount)
}

func TestApplyLibraryEnablementDisablesOmittedLibraries(t *testing.T) {
	const basePath = "/api/v1/settings/plex/library"
	srv := newLibrarySettingsServer(t, basePath, []mockLibrary{
		{ID: "1", Name: "Movies", Enabled: true},
		{ID: "2", Name: "TV", Enabled: true},
	})
	client := testAPIClient(t, srv.URL)

	body, err := applyLibraryEnablement(context.Background(), client, basePath, nil)
	require.NoError(t, err)

	var got []mockLibrary
	require.NoError(t, json.Unmarshal(body, &got))
	assert.Equal(t, []mockLibrary{
		{ID: "1", Name: "Movies", Enabled: false},
		{ID: "2", Name: "TV", Enabled: false},
	}, got)
}

func TestApplyLibraryEnablementRejectsUnknownIDs(t *testing.T) {
	const basePath = "/api/v1/settings/emby/library"
	srv := newLibrarySettingsServer(t, basePath, []mockLibrary{
		{ID: "101", Name: "Movies", Enabled: false},
	})
	client := testAPIClient(t, srv.URL)

	_, err := applyLibraryEnablement(context.Background(), client, basePath, []string{"missing"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown library IDs")
	assert.Equal(t, []mockLibrary{
		{ID: "101", Name: "Movies", Enabled: false},
	}, srv.snapshot())
}

func TestApplyLibraryEnablementIsNoopWhenAlreadyEnabled(t *testing.T) {
	const basePath = "/api/v1/settings/jellyfin/library"
	srv := newLibrarySettingsServer(t, basePath, []mockLibrary{
		{ID: "jf1", Name: "Movies", Enabled: true},
		{ID: "jf2", Name: "TV", Enabled: false},
	})
	client := testAPIClient(t, srv.URL)

	_, err := applyLibraryEnablement(context.Background(), client, basePath, []string{"jf1"})
	require.NoError(t, err)

	for _, call := range srv.recordedCalls() {
		if call.Method == http.MethodPut {
			assert.True(t, strings.HasSuffix(call.Path, "/"+libraryWriteProbeID), "unexpected PUT %s", call.Path)
		}
	}
}

func newLibrarySettingsServerV341(t *testing.T, basePath string, libs []mockLibrary) *librarySettingsServer {
	t.Helper()
	parent := strings.TrimSuffix(basePath, "/library")
	srv := &librarySettingsServer{libs: append([]mockLibrary(nil), libs...)}
	srv.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		srv.mu.Lock()
		defer srv.mu.Unlock()
		srv.calls = append(srv.calls, libraryAPICall{
			Method: r.Method,
			Path:   r.URL.Path,
			Query:  r.URL.Query(),
			Body:   string(body),
		})
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodPut {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"not found","errors":[{"path":"` + r.URL.Path + `","message":"not found"}]}`))
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == parent {
			_ = json.NewEncoder(w).Encode(map[string]any{"libraries": srv.libs})
			return
		}
		if r.Method == http.MethodGet && r.URL.Path == basePath {
			enabledQuery, hasEnable := r.URL.Query()["enable"]
			// Seerr's OpenAPI middleware rejects explicit empty query values.
			if hasEnable && enabledQuery[0] == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message":"Empty value found for query parameter 'enable'"}`))
				return
			}
			enabled := map[string]struct{}{}
			if hasEnable && enabledQuery[0] != "" {
				for _, id := range strings.Split(enabledQuery[0], ",") {
					enabled[id] = struct{}{}
				}
			}
			for i := range srv.libs {
				_, srv.libs[i].Enabled = enabled[srv.libs[i].ID]
			}
			b, _ := json.Marshal(srv.libs)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestApplyLibraryEnablementOnSeerr341UsesGetEnable(t *testing.T) {
	const basePath = "/api/v1/settings/jellyfin/library"
	srv := newLibrarySettingsServerV341(t, basePath, []mockLibrary{
		{ID: "f137a2dd21bbc1b99aa5c0f6bf02a805", Name: "Movies", Enabled: false},
		{ID: "4514ec850e5ad0c47b58444e17b6346c", Name: "TV", Enabled: false},
	})
	client := testAPIClient(t, srv.URL)
	ctx := context.Background()

	body, err := applyLibraryEnablement(ctx, client, basePath, []string{"f137a2dd21bbc1b99aa5c0f6bf02a805", "4514ec850e5ad0c47b58444e17b6346c"})
	require.NoError(t, err)

	var afterUpdate []mockLibrary
	require.NoError(t, json.Unmarshal(body, &afterUpdate))
	assert.Equal(t, []mockLibrary{
		{ID: "f137a2dd21bbc1b99aa5c0f6bf02a805", Name: "Movies", Enabled: true},
		{ID: "4514ec850e5ad0c47b58444e17b6346c", Name: "TV", Enabled: true},
	}, afterUpdate)

	readBody, err := fetchLibraryList(ctx, client, basePath, false, []string{"f137a2dd21bbc1b99aa5c0f6bf02a805", "4514ec850e5ad0c47b58444e17b6346c"})
	require.NoError(t, err)
	var afterRead []mockLibrary
	require.NoError(t, json.Unmarshal(readBody, &afterRead))
	assert.Equal(t, afterUpdate, afterRead)
	assert.Equal(t, afterUpdate, srv.snapshot())

	var wroteViaGetEnable bool
	for _, call := range srv.recordedCalls() {
		if call.Method == http.MethodPut && !strings.Contains(call.Path, libraryWriteProbeID) {
			t.Fatalf("3.4.1 must not PUT real library IDs: %+v", call)
		}
		if call.Method == http.MethodGet && call.Path == basePath && call.Query.Get("enable") != "" {
			wroteViaGetEnable = true
		}
	}
	assert.True(t, wroteViaGetEnable)
}

func TestFetchLibraryListOnSeerr341DoesNotWipeEnabledLibraries(t *testing.T) {
	const basePath = "/api/v1/settings/jellyfin/library"
	srv := newLibrarySettingsServerV341(t, basePath, []mockLibrary{
		{ID: "jf1", Name: "Movies", Enabled: true},
		{ID: "jf2", Name: "TV", Enabled: false},
	})
	client := testAPIClient(t, srv.URL)

	body, err := fetchLibraryList(context.Background(), client, basePath, false, []string{"jf1"})
	require.NoError(t, err)
	var got []mockLibrary
	require.NoError(t, json.Unmarshal(body, &got))
	assert.Equal(t, []mockLibrary{
		{ID: "jf1", Name: "Movies", Enabled: true},
		{ID: "jf2", Name: "TV", Enabled: false},
	}, got)
	assert.Equal(t, got, srv.snapshot())

	for _, call := range srv.recordedCalls() {
		if call.Path == basePath {
			t.Fatalf("3.4.1 read must not GET /library without enable (wipes flags): %+v", call)
		}
	}
}

func TestDetectLibraryWriteModePrefersPutWhenRouteExists(t *testing.T) {
	const basePath = "/api/v1/settings/jellyfin/library"
	srv := newLibrarySettingsServer(t, basePath, nil)
	client := testAPIClient(t, srv.URL)
	mode, err := detectLibraryWriteMode(context.Background(), client, basePath)
	require.NoError(t, err)
	assert.Equal(t, libraryWritePut, mode)
}

func TestDetectLibraryWriteModeFallsBackWhenPutUnmatched(t *testing.T) {
	const basePath = "/api/v1/settings/jellyfin/library"
	srv := newLibrarySettingsServerV341(t, basePath, nil)
	client := testAPIClient(t, srv.URL)
	mode, err := detectLibraryWriteMode(context.Background(), client, basePath)
	require.NoError(t, err)
	assert.Equal(t, libraryWriteGetEnable, mode)
}

func TestDisableAllLibrariesOnSeerr341(t *testing.T) {
	for _, server := range []string{"jellyfin", "plex", "emby"} {
		t.Run(server, func(t *testing.T) {
			basePath := "/api/v1/settings/" + server + "/library"
			srv := newLibrarySettingsServerV341(t, basePath, []mockLibrary{{ID: "movies", Name: "Movies", Enabled: true}})
			body, err := applyLibraryEnablement(context.Background(), testAPIClient(t, srv.URL), basePath, []string{})
			require.NoError(t, err)
			var got []mockLibrary
			require.NoError(t, json.Unmarshal(body, &got))
			require.Len(t, got, 1)
			assert.False(t, got[0].Enabled)
			assert.Equal(t, got, srv.snapshot())
			for _, call := range srv.recordedCalls() {
				if call.Path == basePath {
					assert.NotContains(t, call.Query, "enable")
				}
			}
		})
	}
}
