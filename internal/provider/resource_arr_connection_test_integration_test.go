package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestARRConnectionTestTerraformProtocol(t *testing.T) {
	t.Setenv("TF_ACC", "1")
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		if req.URL.Path != "/api/v1/settings/radarr/test" && req.URL.Path != "/api/v1/settings/sonarr/test" {
			http.NotFound(w, req)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body["apiKey"] != "fake-arr-key" || body["hostname"] == "" || body["port"] == nil || body["useSsl"] == nil {
			http.Error(w, "missing connection fields", http.StatusBadRequest)
			return
		}
		calls = append(calls, req.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"profiles":[{"id":1,"name":"HD"}],"rootFolders":[{"id":2,"path":"/media"}],"tags":[]}`))
	}))
	defer srv.Close()
	config := func(service, hostname string) string {
		return fmt.Sprintf(`provider "seerr" {
  url     = %q
  api_key = "fake-provider-key"
}

resource "seerr_arr_connection_test" "test" {
  service  = %q
  hostname = %q
  port     = 7878
  api_key  = "fake-arr-key"
  use_ssl  = false
  base_url = "/arr"
}`, srv.URL, service, hostname)
	}
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: config("radarr", "radarr.local"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("seerr_arr_connection_test.test", "service", "radarr"), resource.TestCheckResourceAttrSet("seerr_arr_connection_test.test", "response_json"))},
			{Config: config("radarr", "radarr-next.local"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("seerr_arr_connection_test.test", "hostname", "radarr-next.local"))},
			{Config: config("sonarr", "sonarr.local"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("seerr_arr_connection_test.test", "service", "sonarr"))},
		},
	})
	if len(calls) != 3 {
		t.Fatalf("expected create, update, and service replacement connection tests; got %v", calls)
	}
	if calls[0] != "/api/v1/settings/radarr/test" || calls[1] != "/api/v1/settings/radarr/test" || calls[2] != "/api/v1/settings/sonarr/test" {
		t.Fatalf("unexpected endpoint sequence: %v", calls)
	}
}
