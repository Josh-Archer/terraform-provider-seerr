package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserNotificationSettingsDataSource(t *testing.T) {
	username := "test_ds_notif_user"
	email := "test_ds_notif@example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserNotificationSettingsDataSourceConfig_Base(username, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("seerr_user.test", "id"),
				),
			},
			{
				Config: testAccUserNotificationSettingsDataSourceConfig_Base(username, email) + testAccUserNotificationSettingsDataSourceConfig_Step1(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_user_notification_settings.test", "email_enabled", "true"),
					resource.TestCheckResourceAttr("data.seerr_user_notification_settings.test", "discord_enabled", "true"),
					resource.TestCheckResourceAttr("data.seerr_user_notification_settings.test", "discord_id", "987654321"),
					resource.TestCheckResourceAttrSet("data.seerr_user_notification_settings.test", "status_code"),
					resource.TestCheckResourceAttrSet("data.seerr_user_notification_settings.test", "response_json"),
				),
			},
		},
	})
}

func testAccUserNotificationSettingsDataSourceConfig_Base(username, email string) string {
	return fmt.Sprintf(`
resource "seerr_user" "test" {
  username    = %[1]q
  email       = %[2]q
  permissions = 0
}

resource "seerr_user_notification_settings" "test" {
  user_id         = seerr_user.test.id
  email_enabled   = true
  discord_enabled = true
  discord_id      = "987654321"
}
`, username, email)
}

func testAccUserNotificationSettingsDataSourceConfig_Step1() string {
	return `
data "seerr_user_notification_settings" "test" {
  user_id = seerr_user.test.id
}
`
}
