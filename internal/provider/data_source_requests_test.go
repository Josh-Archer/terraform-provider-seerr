package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestRequestMediaTypeFromMedia(t *testing.T) {
	mediaType := requestMediaTypeFromMedia(map[string]any{
		"id":        float64(9),
		"tmdbId":    float64(550),
		"mediaType": "movie",
	})
	if mediaType.IsNull() || mediaType.ValueString() != "movie" {
		t.Fatalf("expected nested mediaType movie, got %v", mediaType)
	}

	if missing := requestMediaTypeFromMedia(map[string]any{"type": "tv"}); !missing.IsNull() {
		t.Fatalf("expected missing nested mediaType to remain null, got %v", missing)
	}
}

// TestRequestsDataSourceMediaTypeAgainstTerraformCLI covers the complete Terraform
// provider path against an API response shaped like Seerr's MediaRequest model.
func TestRequestsDataSourceMediaTypeAgainstTerraformCLI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/request" {
			http.Error(w, "unexpected request", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"id":41,"status":2,"media":{"id":9,"tmdbId":550,"mediaType":"movie"},"requestedBy":{"id":7}}],"pageInfo":{"pages":1,"page":1,"pageSize":100,"results":1,"total":1}}`))
	}))
	defer server.Close()

	config := fmt.Sprintf(`
provider "seerr" {
  url = %q
  api_key = "test-key"
  max_retries = 0
}

data "seerr_requests" "filtered" {
  filter_media_type = "movie"
}
`, server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{{
			Config: config,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.seerr_requests.filtered", "total", "1"),
				resource.TestCheckResourceAttr("data.seerr_requests.filtered", "requests.0.media_type", "movie"),
			),
		}},
	})
}

func TestAccRequestsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "seerr_requests" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_requests.all", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_requests.all", "requests.#"),
				),
			},
			{
				Config: `
data "seerr_requests" "filtered" {
  filter_status = 1
  filter_media_type = "movie"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_requests.filtered", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_requests.filtered", "requests.#"),
					resource.TestCheckResourceAttrSet("data.seerr_requests.filtered", "total"),
				),
			},
		},
	})
}
