package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestMediaStatusTerraformProtocol(t *testing.T) {
	// This local-mock protocol test runs in ordinary go test CI, while the
	// plugin-testing helper still honors TF_ACC_TERRAFORM_PATH when supplied.
	t.Setenv("TF_ACC", "1")
	status := int64(2)
	status4k := int64(1)
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/api/v1/media/7/pending" || req.URL.Path == "/api/v1/media/7/processing" {
			if req.Method != http.MethodPost {
				http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				return
			}
			var body struct {
				Is4K bool `json:"is4k"`
			}
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			if body.Is4K {
				if strings.HasSuffix(req.URL.Path, "/processing") {
					status4k = 3
				} else {
					status4k = 2
				}
			} else if strings.HasSuffix(req.URL.Path, "/processing") {
				status = 3
			} else {
				status = 2
			}
			posts++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 7, "status": status, "status4k": status4k})
			return
		}
		if req.URL.Path == "/api/v1/media/7" && req.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": 7, "status": status, "status4k": status4k})
			return
		}
		http.NotFound(w, req)
	}))
	defer srv.Close()
	config := func(mediaStatus, is4k string) string {
		return fmt.Sprintf(`provider "seerr" {
  url     = %q
  api_key = "fake-test-key"
}

resource "seerr_media_status" "test" {
  media_id = "7"
  status   = %q
  is4k     = %s
}`, srv.URL, mediaStatus, is4k)
	}
	invalidIDConfig := strings.Replace(config("pending", "false"), `media_id = "7"`, `media_id = "not-a-number"`, 1)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{Config: invalidIDConfig, ExpectError: regexp.MustCompile("must be a positive Seerr media ID")},
			{Config: config("pending", "false"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("seerr_media_status.test", "status", "pending"), resource.TestCheckResourceAttr("seerr_media_status.test", "id", "7"))},
			{Config: config("processing", "true"), Check: resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr("seerr_media_status.test", "status", "processing"), resource.TestCheckResourceAttr("seerr_media_status.test", "is4k", "true"))},
			{ResourceName: "seerr_media_status.test", ImportState: true, ImportStateVerify: true, ImportStateId: "7:4k"},
		},
	})
	if posts != 2 {
		t.Fatalf("expected create and update to POST twice, got %d", posts)
	}
}
