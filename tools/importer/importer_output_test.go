// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeCanary builds obviously-fake test values at runtime so secret scanners
// do not treat fixture literals as credentials.
func fakeCanary(name string) string {
	return strings.Join([]string{"canary", name, "not-real"}, "-")
}

func newOutputTestServer(t *testing.T, radarrKey string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/settings/main":
			_ = json.NewEncoder(w).Encode(map[string]any{"applicationTitle": "Seerr"})
		case "/api/v1/settings/radarr":
			_ = json.NewEncoder(w).Encode([]any{map[string]any{
				"id": 1, "name": "Radarr", "hostname": "radarr", "port": 7878,
				"apiKey": radarrKey, "activeProfileId": 1, "activeProfileName": "Any",
				"activeDirectory": "/movies", "minimumAvailability": "released",
			}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestOutputPermissionsAndOverwriteControl(t *testing.T) {
	radarrKey := fakeCanary("radarr-key")
	server := newOutputTestServer(t, radarrKey)
	defer server.Close()

	outDir := filepath.Join(t.TempDir(), "secure-out")
	args := []string{"-url", server.URL, "-api-key", fakeCanary("importer-key"), "-out-dir", outDir, "-format", "all"}

	var stdout, stderr bytes.Buffer
	if code := run(args, &stdout, &stderr); code != 0 {
		t.Fatalf("first run exit %d; stderr:\n%s", code, stderr.String())
	}

	if info, err := os.Stat(outDir); err != nil {
		t.Fatal(err)
	} else if perm := info.Mode().Perm(); perm != 0o700 {
		t.Errorf("out-dir mode = %04o, want 0700", perm)
	}
	want := map[string]os.FileMode{"main.tf": 0o600, "variables.tf": 0o600, "imports.tf": 0o600, "import.sh": 0o700}
	for name, mode := range want {
		info, err := os.Stat(filepath.Join(outDir, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if perm := info.Mode().Perm(); perm != mode {
			t.Errorf("%s mode = %04o, want %04o", name, perm, mode)
		}
		content, err := os.ReadFile(filepath.Join(outDir, name))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(content), radarrKey) {
			t.Errorf("fetched credential leaked into %s", name)
		}
	}
	if strings.Contains(stdout.String()+stderr.String(), radarrKey) {
		t.Errorf("fetched credential leaked into CLI output")
	}

	// A second run without -overwrite must refuse and leave files untouched.
	mainPath := filepath.Join(outDir, "main.tf")
	if err := os.WriteFile(mainPath, []byte("# edited\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(args, &stdout, &stderr); code != 1 {
		t.Errorf("rerun without -overwrite exit %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "refusing to overwrite existing file") {
		t.Errorf("missing refusal message; stderr:\n%s", stderr.String())
	}
	if b, _ := os.ReadFile(mainPath); string(b) != "# edited\n" {
		t.Errorf("main.tf was modified despite refusal")
	}

	// -overwrite replaces the files and resets loose permissions.
	if err := os.Chmod(mainPath, 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := run(append(args, "-overwrite"), &stdout, &stderr); code != 0 {
		t.Fatalf("rerun with -overwrite exit %d; stderr:\n%s", code, stderr.String())
	}
	if info, _ := os.Stat(mainPath); info.Mode().Perm() != 0o600 {
		t.Errorf("main.tf mode after overwrite = %04o, want 0600", info.Mode().Perm())
	}
}

func TestCLIAPIErrorPathDoesNotLeakCredentials(t *testing.T) {
	apiKey := fakeCanary("api-key-auth-fail")
	echoed := fakeCanary("error-echoed-token")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]any{"message": "bad key " + apiKey, "token": echoed})
	}))
	defer server.Close()

	var stdout, stderr bytes.Buffer
	if code := run([]string{"-url", server.URL, "-api-key", apiKey, "-out-dir", t.TempDir()}, &stdout, &stderr); code == 0 {
		t.Fatal("expected nonzero exit on server error")
	}
	out := stdout.String() + stderr.String()
	for _, s := range []string{apiKey, echoed} {
		if strings.Contains(out, s) {
			t.Errorf("credential leaked into CLI output:\n%s", out)
		}
	}
}
