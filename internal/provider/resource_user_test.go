package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("seerr_user", &resource.Sweeper{
		Name: "seerr_user",
		F:    sweepUser,
	})
}

func sweepUser(region string) error {
	client, err := testAccClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	res, err := client.Request(ctx, "GET", "/api/v1/user?take=1000", "", nil)
	if err != nil {
		return fmt.Errorf("error fetching users: %s", err)
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("error fetching users: status %d", res.StatusCode)
	}

	var parsedResponse struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(res.Body, &parsedResponse); err != nil {
		return fmt.Errorf("error parsing users response: %s", err)
	}

	for _, user := range parsedResponse.Results {
		idRaw, ok := user["id"]
		if !ok {
			continue
		}

		var id string
		switch v := idRaw.(type) {
		case float64:
			id = fmt.Sprintf("%.0f", v)
		case string:
			id = v
		}

		if id == "1" {
			log.Printf("[INFO][SWEEPER] Skipping user with ID 1 (default admin)")
			continue
		}

		log.Printf("[INFO][SWEEPER] Deleting user %s", id)
		delRes, err := client.Request(ctx, "DELETE", "/api/v1/user/"+id, "", nil)
		if err != nil {
			log.Printf("[ERROR][SWEEPER] Error deleting user %s: %s", id, err)
			continue
		}
		if !StatusIsOK(delRes.StatusCode) {
			log.Printf("[ERROR][SWEEPER] Error deleting user %s: status %d", id, delRes.StatusCode)
		}
	}

	return nil
}

func TestAccUserResource(t *testing.T) {
	username := "terraform_test_user"
	email := "test_user@example.com"
	updatedUsername := "terraform_test_user_updated"

	// The Plex ID test is difficult to fully automate without a known, stable Plex server attached to the test instance.
	// In a real testing environment, you would provide a valid Plex ID.
	// plexIDToImport := "123456"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create Local User and Verify
			{
				Config: testAccUserResourceConfig(username, email, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user.test", "username", username),
					resource.TestCheckResourceAttr("seerr_user.test", "email", email),
					resource.TestCheckResourceAttr("seerr_user.test", "permissions", "0"),
					resource.TestCheckResourceAttrSet("seerr_user.test", "id"),
				),
			},
			// Update and Verify
			{
				Config: testAccUserResourceConfig(updatedUsername, email, 32), // 32 is MANAGE_SETTINGS
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user.test", "username", updatedUsername),
					resource.TestCheckResourceAttr("seerr_user.test", "permissions", "32"),
				),
			},
			// Update Full Settings
			{
				Config: testAccUserResourceConfigFull(updatedUsername, email, 32),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user.test", "locale", "en"),
					resource.TestCheckResourceAttr("seerr_user.test", "discover_region", "US"),
					resource.TestCheckResourceAttr("seerr_user.test", "notification_settings.telegram_message_thread_id", "1"),
					resource.TestCheckResourceAttr("seerr_user.test", "notification_settings.notification_types.webpush", "256"),
				),
			},
			// Import by ID
			{
				ResourceName:            "seerr_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"notification_settings.pushbullet_access_token", "notification_settings.pushover_application_token", "notification_settings.pushover_user_key"},
			},
			// Import by Username
			{
				ResourceName:            "seerr_user.test",
				ImportState:             true,
				ImportStateId:           updatedUsername, // Importing by username instead of ID
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"notification_settings.pushbullet_access_token", "notification_settings.pushover_application_token", "notification_settings.pushover_user_key"},
			},
			// Import by Email
			{
				ResourceName:            "seerr_user.test",
				ImportState:             true,
				ImportStateId:           email, // Importing by email instead of ID
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"notification_settings.pushbullet_access_token", "notification_settings.pushover_application_token", "notification_settings.pushover_user_key"},
			},
		},
	})
}

func testAccUserResourceConfigFull(username, email string, permissions int) string {
	return providerConfig + fmt.Sprintf(`
resource "seerr_user" "test" {
  username          = %[1]q
  email             = %[2]q
  permissions       = %[3]d
  locale            = "en"
  discover_region   = "US"
  streaming_region  = "US"
  original_language = "en"
  watchlist_sync_movies = true
  watchlist_sync_tv     = true

  notification_settings {
    discord_enabled            = true
    discord_id                 = "123456789"
    telegram_message_thread_id = "1"
    
    notification_types {
      discord    = 2
      webpush    = 256
    }
  }
}
`, username, email, permissions)
}

func testAccUserResourceConfig(username, email string, permissions int) string {
	return providerConfig + fmt.Sprintf(`
resource "seerr_user" "test" {
  username    = %[1]q
  email       = %[2]q
  permissions = %[3]d
}
`, username, email, permissions)
}
