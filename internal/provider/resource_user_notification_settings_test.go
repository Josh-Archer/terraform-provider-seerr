package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestUserNotificationSettingsPath(t *testing.T) {
	if got, want := userNotificationSettingsPath(42), "/api/v1/user/42/settings/notifications"; got != want {
		t.Fatalf("userNotificationSettingsPath(42) = %q, want %q", got, want)
	}
}

func TestPopulateUserNotificationSettingsFromMap(t *testing.T) {
	decoded := map[string]any{
		"emailEnabled":        true,
		"pgpKey":              "test-pgp-key",
		"discordEnabled":      true,
		"discordId":           "999888777",
		"telegramEnabled":     true,
		"telegramBotUsername": "testbot",
		"telegramChatId":      "123456",
		"notificationTypes": map[string]any{
			"discord": float64(2),
			"email":   float64(4),
		},
	}

	var (
		emailEnabled             types.Bool
		pgpKey                   types.String
		discordEnabled           types.Bool
		discordID                types.String
		pushbulletAccessToken    types.String
		pushoverApplicationToken types.String
		pushoverUserKey          types.String
		pushoverSound            types.String
		telegramEnabled          types.Bool
		telegramBotUsername      types.String
		telegramChatID           types.String
		telegramMessageThreadID  types.String
		telegramSendSilently     types.Bool
		webpushEnabled           types.Bool
		notificationTypes        *UserNotificationTypesModel
	)

	populateUserNotificationSettingsFromMap(
		decoded,
		&emailEnabled,
		&pgpKey,
		&discordEnabled,
		&discordID,
		&pushbulletAccessToken,
		&pushoverApplicationToken,
		&pushoverUserKey,
		&pushoverSound,
		&telegramEnabled,
		&telegramBotUsername,
		&telegramChatID,
		&telegramMessageThreadID,
		&telegramSendSilently,
		&webpushEnabled,
		&notificationTypes,
	)

	if !emailEnabled.ValueBool() {
		t.Errorf("expected email_enabled true")
	}
	if pgpKey.ValueString() != "test-pgp-key" {
		t.Errorf("pgpKey = %q, want test-pgp-key", pgpKey.ValueString())
	}
	if !discordEnabled.ValueBool() {
		t.Errorf("expected discord_enabled true")
	}
	if discordID.ValueString() != "999888777" {
		t.Errorf("discordID = %q, want 999888777", discordID.ValueString())
	}
	if notificationTypes == nil || notificationTypes.Discord.ValueInt64() != 2 || notificationTypes.Email.ValueInt64() != 4 {
		t.Errorf("notificationTypes bitmask not correctly populated")
	}
}

func TestAccUserNotificationSettingsResource(t *testing.T) {
	username := "test_notif_user"
	email := "test_notif@example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserNotificationSettingsConfig_Base(username, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("seerr_user.test", "id"),
				),
			},
			{
				Config: testAccUserNotificationSettingsConfig_Base(username, email) + testAccUserNotificationSettingsConfig_Step1(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user_notification_settings.test", "email_enabled", "true"),
					resource.TestCheckResourceAttr("seerr_user_notification_settings.test", "discord_enabled", "true"),
					resource.TestCheckResourceAttr("seerr_user_notification_settings.test", "discord_id", "123456789"),
					resource.TestCheckResourceAttr("seerr_user_notification_settings.test", "notification_types.discord", "2"),
				),
			},
			{
				Config: testAccUserNotificationSettingsConfig_Base(username, email) + testAccUserNotificationSettingsConfig_Step2(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user_notification_settings.test", "email_enabled", "false"),
					resource.TestCheckResourceAttr("seerr_user_notification_settings.test", "discord_enabled", "false"),
				),
			},
			{
				ResourceName:      "seerr_user_notification_settings.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserNotificationSettingsConfig_Base(username, email string) string {
	return fmt.Sprintf(`
resource "seerr_user" "test" {
  username    = %[1]q
  email       = %[2]q
  permissions = 0
}
`, username, email)
}

func testAccUserNotificationSettingsConfig_Step1() string {
	return `
resource "seerr_user_notification_settings" "test" {
  user_id         = seerr_user.test.id
  email_enabled   = true
  discord_enabled = true
  discord_id      = "123456789"
  notification_types = {
    discord = 2
  }
}
`
}

func testAccUserNotificationSettingsConfig_Step2() string {
	return `
resource "seerr_user_notification_settings" "test" {
  user_id         = seerr_user.test.id
  email_enabled   = false
  discord_enabled = false
}
`
}
