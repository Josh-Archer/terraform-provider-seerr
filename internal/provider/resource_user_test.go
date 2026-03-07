package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

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
			// Update Notification Settings
			{
				Config: testAccUserResourceConfigWithNotifications(updatedUsername, email, 32),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user.test", "notification_settings.discord_enabled", "true"),
					resource.TestCheckResourceAttr("seerr_user.test", "notification_settings.discord_id", "123456789"),
					resource.TestCheckResourceAttr("seerr_user.test", "notification_settings.notification_types.discord", "2"),
				),
			},
			// Import
			{
				ResourceName:      "seerr_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserResourceConfig(username, email string, permissions int) string {
	return fmt.Sprintf(`
resource "seerr_user" "test" {
  username    = %[1]q
  email       = %[2]q
  permissions = %[3]d
}
`, username, email, permissions)
}

func testAccUserResourceConfigWithNotifications(username, email string, permissions int) string {
	return fmt.Sprintf(`
resource "seerr_user" "test" {
  username    = %[1]q
  email       = %[2]q
  permissions = %[3]d

  notification_settings = {
    discord_enabled = true
    discord_id      = "123456789"
    
    notification_types = {
      discord = 2
    }
  }
}
`, username, email, permissions)
}
