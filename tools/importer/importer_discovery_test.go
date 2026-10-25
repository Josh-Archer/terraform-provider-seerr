// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestAssignUniqueNamesIsDeterministic(t *testing.T) {
	resources := []DiscoveredResource{
		{ResourceType: "seerr_radarr_server", ResourceName: "radarr_hd", ImportID: "5"},
		{ResourceType: "seerr_radarr_server", ResourceName: "radarr_hd", ImportID: "1"},
		{ResourceType: "seerr_user", ResourceName: "admin", ImportID: "10"},
		{ResourceType: "seerr_user", ResourceName: "admin", ImportID: "2"},
		{ResourceType: "seerr_user", ResourceName: "admin", ImportID: "3"},
		{ResourceType: "seerr_user", ResourceName: "admin_2", ImportID: "4"},
	}
	assignUniqueResourceNames(resources)

	got := map[string]string{}
	for _, resource := range resources {
		got[resource.ResourceType+"/"+resource.ImportID] = resource.ResourceName
	}
	want := map[string]string{
		"seerr_radarr_server/1": "radarr_hd",
		"seerr_radarr_server/5": "radarr_hd_2",
		"seerr_user/2":          "admin",
		"seerr_user/3":          "admin_2",
		"seerr_user/4":          "admin_2_2",
		"seerr_user/10":         "admin_3",
	}
	for key, name := range want {
		if got[key] != name {
			t.Errorf("%s name = %q, want %q", key, got[key], name)
		}
	}
}

func TestDiscoveryNamesThatSanitizeTogether(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/radarr" {
			http.NotFound(w, r)
			return
		}
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 7, "name": "Radarr HD", "hostname": "a.local", "port": 7878, "apiKey": "secret-a", "activeProfileId": 1, "activeDirectory": "/a"},
			{"id": 2, "name": "radarr-hd", "hostname": "b.local", "port": 7878, "apiKey": "secret-b", "activeProfileId": 1, "activeDirectory": "/b"},
			{"id": 3, "name": "radarr.hd", "hostname": "c.local", "port": 7878, "apiKey": "secret-c", "activeProfileId": 1, "activeDirectory": "/c"},
		})
	}))
	defer server.Close()

	imp := NewImporter(ImporterConfig{BaseURL: server.URL, HTTPClient: server.Client()})
	resources, err := imp.DiscoverAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]string{}
	for _, resource := range resources {
		if resource.ResourceType == "seerr_radarr_server" {
			names[resource.ImportID] = resource.ResourceName
		}
	}
	if names["2"] != "radarr_hd" || names["3"] != "radarr_hd_2" || names["7"] != "radarr_hd_3" {
		t.Fatalf("names by import id = %#v", names)
	}
	hcl := GenerateHCL(resources, false)
	for _, secret := range []string{"secret-a", "secret-b", "secret-c"} {
		if strings.Contains(hcl, secret) {
			t.Fatalf("secret %s written into HCL:\n%s", secret, hcl)
		}
	}
}

func TestUserPaginationBoundaries(t *testing.T) {
	t.Run("more than one page", func(t *testing.T) {
		requests := serveUsers(t, 101, false)
		imp := NewImporter(ImporterConfig{BaseURL: requests.url, HTTPClient: requests.client})
		resources, err := imp.DiscoverAll(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got := countType(resources, "seerr_user"); got != 101 {
			t.Fatalf("users = %d, want 101", got)
		}
		if strings.Join(requests.skips(), ",") != "0,100" {
			t.Fatalf("skips = %v", requests.skips())
		}
	})

	t.Run("exact page of 100", func(t *testing.T) {
		requests := serveUsers(t, 100, false)
		imp := NewImporter(ImporterConfig{BaseURL: requests.url, HTTPClient: requests.client})
		resources, err := imp.DiscoverAll(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got := countType(resources, "seerr_user"); got != 100 {
			t.Fatalf("users = %d, want 100", got)
		}
		if strings.Join(requests.skips(), ",") != "0,100" {
			t.Fatalf("skips = %v, want a following empty page", requests.skips())
		}
	})

	t.Run("exact two pages", func(t *testing.T) {
		requests := serveUsers(t, 200, false)
		imp := NewImporter(ImporterConfig{BaseURL: requests.url, HTTPClient: requests.client})
		resources, err := imp.DiscoverAll(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got := countType(resources, "seerr_user"); got != 200 {
			t.Fatalf("users = %d, want 200", got)
		}
		if strings.Join(requests.skips(), ",") != "0,100,200" {
			t.Fatalf("skips = %v", requests.skips())
		}
	})

	t.Run("pageInfo total stops on the boundary", func(t *testing.T) {
		requests := serveUsers(t, 100, true)
		imp := NewImporter(ImporterConfig{BaseURL: requests.url, HTTPClient: requests.client})
		resources, err := imp.DiscoverAll(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if got := countType(resources, "seerr_user"); got != 100 {
			t.Fatalf("users = %d, want 100", got)
		}
		if strings.Join(requests.skips(), ",") != "0" {
			t.Fatalf("skips = %v, want a single page when total is reached", requests.skips())
		}
	})
}

func TestUserPaginationDedupesAndStops(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user" {
			http.NotFound(w, r)
			return
		}
		calls = append(calls, r.URL.Query().Get("skip"))
		users := make([]map[string]any, 100)
		for i := range users {
			users[i] = map[string]any{"id": i + 1, "username": "user", "email": "user@example.com", "permissions": 0}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"results": users})
	}))
	defer server.Close()

	imp := NewImporter(ImporterConfig{BaseURL: server.URL, HTTPClient: server.Client(), MaxUserPages: 20})
	resources, err := imp.DiscoverAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := countType(resources, "seerr_user"); got != 100 {
		t.Fatalf("users = %d, want 100 after dedupe", got)
	}
	if strings.Join(calls, ",") != "0,100" {
		t.Fatalf("pagination did not stop on a repeated page: %v", calls)
	}
	if resources[0].ResourceName == "" {
		t.Fatal("missing resource name")
	}
}

func TestUserPaginationIsBounded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user" {
			http.NotFound(w, r)
			return
		}
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		take, _ := strconv.Atoi(r.URL.Query().Get("take"))
		if take != 2 {
			t.Errorf("take = %d, want 2", take)
		}
		users := []map[string]any{
			{"id": skip + 1, "username": "a", "email": "a@example.com"},
			{"id": skip + 2, "username": "b", "email": "b@example.com"},
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"results": users})
	}))
	defer server.Close()

	imp := NewImporter(ImporterConfig{
		BaseURL:      server.URL,
		HTTPClient:   server.Client(),
		UserPageSize: 2,
		MaxUserPages: 2,
	})
	if _, err := imp.DiscoverAll(context.Background()); err == nil || !strings.Contains(err.Error(), "pagination") {
		t.Fatalf("expected bounded pagination error, got %v", err)
	}

	partial := NewImporter(ImporterConfig{
		BaseURL:      server.URL,
		HTTPClient:   server.Client(),
		UserPageSize: 2,
		MaxUserPages: 2,
		AllowPartial: true,
	})
	resources, err := partial.DiscoverAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got := countType(resources, "seerr_user"); got != 4 {
		t.Fatalf("partial users = %d, want 4", got)
	}
	if len(partial.Failures()) != 1 {
		t.Fatalf("failures = %v", partial.Failures())
	}
}

func TestDiscoveryClassifiesEndpointErrors(t *testing.T) {
	canary := "error-body-canary-should-not-leak"
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusBadRequest, http.StatusInternalServerError} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/api/v1/settings/main" {
					_ = json.NewEncoder(w).Encode(map[string]any{"applicationTitle": "Title"})
					return
				}
				if r.URL.Path == "/api/v1/settings/radarr" {
					w.WriteHeader(status)
					_, _ = w.Write([]byte(canary))
					return
				}
				http.NotFound(w, r)
			}))
			defer server.Close()

			imp := NewImporter(ImporterConfig{BaseURL: server.URL, HTTPClient: server.Client()})
			_, err := imp.DiscoverAll(context.Background())
			if err == nil {
				t.Fatal("expected discovery to fail")
			}
			if strings.Contains(err.Error(), canary) {
				t.Fatalf("error included response body: %v", err)
			}
			message := err.Error()
			switch status {
			case http.StatusUnauthorized, http.StatusForbidden:
				if !strings.Contains(message, "authentication error") {
					t.Fatalf("message = %s", message)
				}
			case http.StatusBadRequest:
				if !strings.Contains(message, "request error") {
					t.Fatalf("message = %s", message)
				}
			default:
				if !strings.Contains(message, "server error") {
					t.Fatalf("message = %s", message)
				}
			}
		})
	}
}

func TestUnsupportedEndpointDoesNotFailDiscovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/main" {
			_ = json.NewEncoder(w).Encode(map[string]any{"applicationTitle": "Title", "locale": "en"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	imp := NewImporter(ImporterConfig{BaseURL: server.URL, HTTPClient: server.Client()})
	resources, err := imp.DiscoverAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if countType(resources, "seerr_main_settings") != 1 {
		t.Fatalf("resources = %#v", resources)
	}
	if len(imp.Failures()) != 0 {
		t.Fatalf("404s were treated as failures: %v", imp.Failures())
	}
	if len(imp.UnsupportedEndpoints()) == 0 {
		t.Fatal("expected unsupported endpoints")
	}
}

func TestTransportErrorFailsDiscovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	imp := NewImporter(ImporterConfig{BaseURL: server.URL, HTTPClient: server.Client()})
	_, err := imp.DiscoverAll(context.Background())
	if err == nil || !strings.Contains(err.Error(), "transport error") {
		t.Fatalf("expected transport error, got %v", err)
	}
}

func TestPopulatedUnmanagedEndpointsAreNotEmitted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/settings/network":
			_ = json.NewEncoder(w).Encode(map[string]any{"trustProxy": true, "proxy": map[string]any{"hostname": "proxy.local"}})
		case "/api/v1/settings/discover":
			_ = json.NewEncoder(w).Encode([]any{map[string]any{"id": 1, "title": "Trending", "type": 1, "enabled": true}})
		case "/api/v1/overrideRule":
			_ = json.NewEncoder(w).Encode([]any{map[string]any{"id": 4, "keywords": "anime"}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	imp := NewImporter(ImporterConfig{BaseURL: server.URL, HTTPClient: server.Client()})
	resources, err := imp.DiscoverAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, resource := range resources {
		switch resource.ResourceType {
		case "seerr_network_settings", "seerr_discover_slider", "seerr_override_rule":
			t.Fatalf("emitted unmanaged resource %s", resource.ResourceType)
		}
	}
	notes := strings.Join(imp.ManualAdoption(), "\n")
	for _, needle := range []string{"seerr_network_settings", "seerr_discover_slider", "seerr_override_rule"} {
		if !strings.Contains(notes, needle) {
			t.Fatalf("manual adoption notes missing %s:\n%s", needle, notes)
		}
	}
}

func TestAllowPartialKeepsSuccessfulResources(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/main" {
			_ = json.NewEncoder(w).Encode(map[string]any{"applicationTitle": "Title"})
			return
		}
		if r.URL.Path == "/api/v1/user" {
			http.Error(w, "nope", http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	strict := NewImporter(ImporterConfig{BaseURL: server.URL, HTTPClient: server.Client()})
	if _, err := strict.DiscoverAll(context.Background()); err == nil {
		t.Fatal("expected strict discovery to fail")
	}

	partial := NewImporter(ImporterConfig{BaseURL: server.URL, HTTPClient: server.Client(), AllowPartial: true})
	resources, err := partial.DiscoverAll(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if countType(resources, "seerr_main_settings") != 1 {
		t.Fatalf("partial resources = %#v", resources)
	}
	if len(partial.Failures()) == 0 {
		t.Fatal("expected recorded failures")
	}
}

func TestCLIValidatesFormatBeforeDiscovery(t *testing.T) {
	hits := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		http.Error(w, "should not be called", http.StatusInternalServerError)
	}))
	defer server.Close()

	dir := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := run([]string{"-url", server.URL, "-format", "yaml", "-out-dir", dir}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("expected nonzero exit")
	}
	if hits != 0 {
		t.Fatalf("discovery ran %d times", hits)
	}
	if strings.Contains(stdout.String(), "Connecting") {
		t.Fatalf("stdout = %s", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(dir, "main.tf")); !os.IsNotExist(err) {
		t.Fatal("main.tf was written for an invalid format")
	}
}

func TestCLIWriteFailureExitsNonzero(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/main" {
			_ = json.NewEncoder(w).Encode(map[string]any{"applicationTitle": "Title"})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	dir := t.TempDir()
	blocker := filepath.Join(dir, "not-a-directory")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := run([]string{"-url", server.URL, "-format", "hcl", "-out-dir", blocker}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("expected write failure, stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "Error creating output directory") {
		t.Fatalf("stderr = %s", stderr.String())
	}
}

func TestCLIPartialModeIsExplicit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/settings/main" {
			_ = json.NewEncoder(w).Encode(map[string]any{"applicationTitle": "Title", "locale": "en"})
			return
		}
		if r.URL.Path == "/api/v1/settings/radarr" {
			http.Error(w, "down", http.StatusInternalServerError)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	strictDir := t.TempDir()
	var strictErr bytes.Buffer
	if code := run([]string{"-url", server.URL, "-format", "hcl", "-out-dir", strictDir}, ioDiscard(), &strictErr); code == 0 {
		t.Fatal("strict discovery exited 0")
	}
	if _, err := os.Stat(filepath.Join(strictDir, "main.tf")); !os.IsNotExist(err) {
		t.Fatal("strict discovery wrote main.tf")
	}
	if !strings.Contains(strictErr.String(), "server error") {
		t.Fatalf("stderr = %s", strictErr.String())
	}

	partialDir := t.TempDir()
	var partialErr bytes.Buffer
	if code := run([]string{"-url", server.URL, "-format", "hcl", "-out-dir", partialDir, "-allow-partial"}, ioDiscard(), &partialErr); code != 0 {
		t.Fatalf("partial discovery failed: %s", partialErr.String())
	}
	if _, err := os.Stat(filepath.Join(partialDir, "main.tf")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(partialErr.String(), "allow-partial") {
		t.Fatalf("stderr = %s", partialErr.String())
	}
}

func TestMigrationGuideDocumentsManualAdoption(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	root := filepath.Join(filepath.Dir(source), "../..")
	guide, err := os.ReadFile(filepath.Join(root, "docs/guides/migration.md"))
	if err != nil {
		t.Fatal(err)
	}
	template, err := os.ReadFile(filepath.Join(root, "templates/guides/migration.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(guide) != string(template) {
		t.Fatal("docs/guides/migration.md and templates/guides/migration.md diverged")
	}
	for _, needle := range []string{
		"seerr_metadata_settings",
		"seerr_discover_slider",
		"seerr_override_rule",
		"seerr_network_settings",
		"allow-partial",
		"terraform.tfvars",
		"take=100",
	} {
		if !strings.Contains(string(guide), needle) {
			t.Errorf("migration guide missing %s", needle)
		}
	}
}

type userPageServer struct {
	url    string
	client *http.Client
	calls  *[]string
}

func (s userPageServer) skips() []string {
	return append([]string(nil), (*s.calls)...)
}

func serveUsers(t *testing.T, total int, includeTotal bool) userPageServer {
	t.Helper()
	calls := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("take") != "100" {
			t.Errorf("take = %s", r.URL.Query().Get("take"))
		}
		skip, _ := strconv.Atoi(r.URL.Query().Get("skip"))
		calls = append(calls, strconv.Itoa(skip))
		end := skip + 100
		if end > total {
			end = total
		}
		var results []map[string]any
		if skip < total {
			for id := skip + 1; id <= end; id++ {
				results = append(results, map[string]any{
					"id":          id,
					"username":    "user" + strconv.Itoa(id),
					"email":       "user" + strconv.Itoa(id) + "@example.com",
					"permissions": 1,
				})
			}
		}
		payload := map[string]any{"results": results}
		if includeTotal {
			pages := (total + 99) / 100
			page := (skip / 100) + 1
			if len(results) == 0 {
				page = pages
			}
			payload["pageInfo"] = map[string]any{"total": total, "pages": pages, "page": page, "pageSize": 100}
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	t.Cleanup(server.Close)
	return userPageServer{url: server.URL, client: server.Client(), calls: &calls}
}

func countType(resources []DiscoveredResource, resourceType string) int {
	count := 0
	for _, resource := range resources {
		if resource.ResourceType == resourceType {
			count++
		}
	}
	return count
}

func ioDiscard() *bytes.Buffer {
	return &bytes.Buffer{}
}
