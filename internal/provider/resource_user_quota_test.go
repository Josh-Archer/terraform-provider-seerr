package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ----------------------------------------------------------------------------
// Unit tests (httptest-based, no live Seerr required)
// ----------------------------------------------------------------------------

func TestApplyQuotaResponseMapsAllFields(t *testing.T) {
	body := []byte(`{
		"movieQuotaLimit":10,
		"movieQuotaDays":30,
		"tvQuotaLimit":5,
		"tvQuotaDays":14,
		"globalMovieQuotaLimit":20,
		"globalMovieQuotaDays":60,
		"globalTvQuotaLimit":8,
		"globalTvQuotaDays":21
	}`)

	data := &UserQuotaModel{UserID: types.Int64Value(7)}
	applyQuotaResponse(data, body)

	if got := data.ID.ValueString(); got != "7" {
		t.Fatalf("expected id '7', got %q", got)
	}
	if got := data.MovieQuotaLimit.ValueInt64(); got != 10 {
		t.Fatalf("expected MovieQuotaLimit 10, got %d", got)
	}
	if got := data.MovieQuotaDays.ValueInt64(); got != 30 {
		t.Fatalf("expected MovieQuotaDays 30, got %d", got)
	}
	if got := data.TvQuotaLimit.ValueInt64(); got != 5 {
		t.Fatalf("expected TvQuotaLimit 5, got %d", got)
	}
	if got := data.TvQuotaDays.ValueInt64(); got != 14 {
		t.Fatalf("expected TvQuotaDays 14, got %d", got)
	}
	if got := data.GlobalMovieQuotaLimit.ValueInt64(); got != 20 {
		t.Fatalf("expected GlobalMovieQuotaLimit 20, got %d", got)
	}
	if got := data.GlobalMovieQuotaDays.ValueInt64(); got != 60 {
		t.Fatalf("expected GlobalMovieQuotaDays 60, got %d", got)
	}
	if got := data.GlobalTvQuotaLimit.ValueInt64(); got != 8 {
		t.Fatalf("expected GlobalTvQuotaLimit 8, got %d", got)
	}
	if got := data.GlobalTvQuotaDays.ValueInt64(); got != 21 {
		t.Fatalf("expected GlobalTvQuotaDays 21, got %d", got)
	}
}

func TestApplyQuotaResponseDefaultsToZeroWhenAbsent(t *testing.T) {
	// Seerr may return a response without quota fields for users that haven't
	// been configured; they should default to 0 (inherit global).
	body := []byte(`{"locale":"en"}`)
	data := &UserQuotaModel{UserID: types.Int64Value(3)}
	applyQuotaResponse(data, body)

	for name, got := range map[string]int64{
		"MovieQuotaLimit": data.MovieQuotaLimit.ValueInt64(),
		"MovieQuotaDays":  data.MovieQuotaDays.ValueInt64(),
		"TvQuotaLimit":    data.TvQuotaLimit.ValueInt64(),
		"TvQuotaDays":     data.TvQuotaDays.ValueInt64(),
	} {
		if got != 0 {
			t.Fatalf("expected %s to default to 0, got %d", name, got)
		}
	}
}

func TestUserQuotaApplyQuotaPreservesNonQuotaFields(t *testing.T) {
	var capturedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/user/5/settings/main":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"locale":"fr",
				"discoverRegion":"FR",
				"movieQuotaLimit":0,
				"movieQuotaDays":0,
				"tvQuotaLimit":0,
				"tvQuotaDays":0
			}`))
		case "POST /api/v1/user/5/settings/main":
			if err := json.NewDecoder(r.Body).Decode(new(map[string]any)); err == nil {
				// capture for inspection
				_ = json.NewDecoder(r.Body)
			}
			// Read the body properly
			var buf map[string]any
			_ = json.NewDecoder(r.Body)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
			_ = buf
			_ = capturedBody
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	// Use a capturing server variant so we can inspect the written payload.
	var written map[string]any
	srvCapture := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/user/5/settings/main":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"locale":"fr",
				"discoverRegion":"FR",
				"movieQuotaLimit":0,
				"movieQuotaDays":0,
				"tvQuotaLimit":0,
				"tvQuotaDays":0
			}`))
		case "POST /api/v1/user/5/settings/main":
			_ = json.NewDecoder(r.Body).Decode(&written)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srvCapture.Close()
	_ = srv

	baseURL, _ := url.Parse(srvCapture.URL)
	res := &UserQuotaResource{client: NewClient(baseURL, "tok", "agent", false, defaultRequestTimeout)}

	data := &UserQuotaModel{
		UserID:          types.Int64Value(5),
		MovieQuotaLimit: types.Int64Value(10),
		MovieQuotaDays:  types.Int64Value(30),
		TvQuotaLimit:    types.Int64Null(), // unset — should not be included
		TvQuotaDays:     types.Int64Null(),
	}

	if err := res.applyQuota(context.Background(), data); err != nil {
		t.Fatalf("applyQuota returned error: %v", err)
	}

	// Non-quota fields from the GET response should be preserved.
	if got, ok := written["locale"].(string); !ok || got != "fr" {
		t.Fatalf("expected locale 'fr' preserved, got %#v", written["locale"])
	}
	if got, ok := written["discoverRegion"].(string); !ok || got != "FR" {
		t.Fatalf("expected discoverRegion 'FR' preserved, got %#v", written["discoverRegion"])
	}
	// The explicit plan values should be written.
	if v, ok := int64ValueFromAny(written["movieQuotaLimit"]); !ok || v != 10 {
		t.Fatalf("expected movieQuotaLimit 10, got %#v", written["movieQuotaLimit"])
	}
	if v, ok := int64ValueFromAny(written["movieQuotaDays"]); !ok || v != 30 {
		t.Fatalf("expected movieQuotaDays 30, got %#v", written["movieQuotaDays"])
	}
}

// ----------------------------------------------------------------------------
// Acceptance tests (require live Seerr; gated by testAccProtoV6ProviderFactories)
// ----------------------------------------------------------------------------

func TestAccUserQuotaResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create a standard_user quota (5 movies / 7 days; 3 TV / 7 days)
			{
				Config: testAccUserQuotaConfig(2, 5, 7, 3, 7),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user_quota.test", "movie_quota_limit", "5"),
					resource.TestCheckResourceAttr("seerr_user_quota.test", "movie_quota_days", "7"),
					resource.TestCheckResourceAttr("seerr_user_quota.test", "tv_quota_limit", "3"),
					resource.TestCheckResourceAttr("seerr_user_quota.test", "tv_quota_days", "7"),
				),
			},
			// Step 2: Upgrade to power_user quota (20 movies / 30 days; 10 TV / 30 days)
			{
				Config: testAccUserQuotaConfig(2, 20, 30, 10, 30),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user_quota.test", "movie_quota_limit", "20"),
					resource.TestCheckResourceAttr("seerr_user_quota.test", "movie_quota_days", "30"),
					resource.TestCheckResourceAttr("seerr_user_quota.test", "tv_quota_limit", "10"),
					resource.TestCheckResourceAttr("seerr_user_quota.test", "tv_quota_days", "30"),
				),
			},
			// Step 3: Reset to unlimited (0 = inherit global)
			{
				Config: testAccUserQuotaConfig(2, 0, 0, 0, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user_quota.test", "movie_quota_limit", "0"),
					resource.TestCheckResourceAttr("seerr_user_quota.test", "tv_quota_limit", "0"),
				),
			},
			// Step 4: Import by user_id
			{
				ResourceName:      "seerr_user_quota.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserQuotaConfig(userID, movieLimit, movieDays, tvLimit, tvDays int) string {
	return providerConfig + fmt.Sprintf(`
resource "seerr_user_quota" "test" {
  user_id           = %d
  movie_quota_limit = %d
  movie_quota_days  = %d
  tv_quota_limit    = %d
  tv_quota_days     = %d
}
`, userID, movieLimit, movieDays, tvLimit, tvDays)
}
