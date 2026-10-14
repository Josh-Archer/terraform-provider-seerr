package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUserRequestsDataSource_Schema(t *testing.T) {
	ds := NewUserRequestsDataSource()
	var resp datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema diagnostics error: %v", resp.Diagnostics)
	}

	attrs := []string{"id", "user_id", "take", "skip", "page_info", "requests"}
	for _, attr := range attrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Expected %q attribute in schema", attr)
		}
	}
}

func TestUserRequestsDataSource_Metadata(t *testing.T) {
	ds := NewUserRequestsDataSource()
	var metaResp datasource.MetadataResponse
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, &metaResp)
	if metaResp.TypeName != "seerr_user_requests" {
		t.Errorf("Expected type name seerr_user_requests, got %s", metaResp.TypeName)
	}
}

func TestUserRequestsDataSource_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "expected GET", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/api/v1/user/10/requests" {
			http.Error(w, fmt.Sprintf("unexpected path %s", r.URL.Path), http.StatusNotFound)
			return
		}
		query := r.URL.Query()
		if query.Get("take") != "20" || query.Get("skip") != "0" {
			http.Error(w, fmt.Sprintf("unexpected query %s", r.URL.RawQuery), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"pageInfo": {
				"page": 1,
				"pages": 1,
				"results": 2
			},
			"results": [
				{
					"id": 101,
					"status": 2,
					"is4k": false,
					"createdAt": "2026-01-01T10:00:00.000Z",
					"updatedAt": "2026-01-01T11:00:00.000Z",
					"media": {
						"id": 50,
						"tmdbId": 603,
						"mediaType": "movie"
					}
				},
				{
					"id": 102,
					"status": 1,
					"is4k": true,
					"createdAt": "2026-01-02T10:00:00.000Z",
					"updatedAt": "2026-01-02T10:00:00.000Z",
					"media": {
						"id": 51,
						"tmdbId": 1399,
						"mediaType": "tv"
					}
				}
			]
		}`))
	}))
	defer server.Close()

	config := fmt.Sprintf(`
provider "seerr" {
  url         = %q
  api_key     = "test-key"
  max_retries = 0
}

data "seerr_user_requests" "test" {
  user_id = 10
  take    = 20
  skip    = 0
}
`, server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "id", "user_10_requests_take_20_skip_0"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "user_id", "10"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "take", "20"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "skip", "0"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "page_info.page", "1"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "page_info.pages", "1"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "page_info.results", "2"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.#", "2"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.0.id", "101"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.0.status", "2"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.0.is_4k", "false"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.0.media_id", "50"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.0.tmdb_id", "603"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.0.media_type", "movie"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.0.created_at", "2026-01-01T10:00:00.000Z"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.0.updated_at", "2026-01-01T11:00:00.000Z"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.1.id", "102"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.1.status", "1"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.1.is_4k", "true"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.1.media_id", "51"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.1.tmdb_id", "1399"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.test", "requests.1.media_type", "tv"),
				),
			},
		},
	})
}

func TestUserRequestsDataSource_EmptyResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/user/15/requests" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"pageInfo": {
				"page": 1,
				"pages": 0,
				"results": 0
			},
			"results": []
		}`))
	}))
	defer server.Close()

	config := fmt.Sprintf(`
provider "seerr" {
  url         = %q
  api_key     = "test-key"
  max_retries = 0
}

data "seerr_user_requests" "empty" {
  user_id = 15
}
`, server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_user_requests.empty", "id", "user_15_requests"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.empty", "user_id", "15"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.empty", "page_info.results", "0"),
					resource.TestCheckResourceAttr("data.seerr_user_requests.empty", "requests.#", "0"),
				),
			},
		},
	})
}

func TestUserRequestsDataSource_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"User not found."}`, http.StatusNotFound)
	}))
	defer server.Close()

	config := fmt.Sprintf(`
provider "seerr" {
  url         = %q
  api_key     = "test-key"
  max_retries = 0
}

data "seerr_user_requests" "missing" {
  user_id = 999
}
`, server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile("user 999 not found"),
			},
		},
	})
}
