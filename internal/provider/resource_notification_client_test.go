package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationNtfyResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationNtfyConfig("https://ntfy.example.com", "terraform-create", "token-create", 3, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "enabled", "true"),
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "embed_poster", "true"),
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "ntfy.url", "https://ntfy.example.com"),
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "ntfy.topic", "terraform-create"),
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "ntfy.priority", "3"),
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "notification_types.#", "2"),
					resource.TestCheckResourceAttrSet("seerr_notification_ntfy.test", "id"),
				),
			},
			{
				Config: testAccNotificationNtfyConfig("https://ntfy.example.com/v2", "terraform-update", "token-update", 5, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "embed_poster", "false"),
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "ntfy.url", "https://ntfy.example.com/v2"),
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "ntfy.topic", "terraform-update"),
					resource.TestCheckResourceAttr("seerr_notification_ntfy.test", "ntfy.priority", "5"),
				),
			},
			{
				ResourceName:            "seerr_notification_ntfy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"ntfy.token"},
			},
			{
				Config: testAccNotificationNtfyConfig("https://ntfy.example.com/v2", "terraform-update", "token-update", 5, false) + testAccNotificationNtfyDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_notification_ntfy.current", "enabled", "true"),
					resource.TestCheckResourceAttr("data.seerr_notification_ntfy.current", "ntfy.url", "https://ntfy.example.com/v2"),
					resource.TestCheckResourceAttr("data.seerr_notification_ntfy.current", "ntfy.topic", "terraform-update"),
					resource.TestCheckResourceAttr("data.seerr_notification_ntfy.current", "ntfy.priority", "5"),
				),
			},
		},
	})
}

func TestAccNotificationPushoverResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNotificationPushoverConfig("app-token-create", "user-token-create", "bike"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_notification_pushover.test", "enabled", "true"),
					resource.TestCheckResourceAttr("seerr_notification_pushover.test", "pushover.sound", "bike"),
					resource.TestCheckResourceAttr("seerr_notification_pushover.test", "notification_types.#", "2"),
					resource.TestCheckResourceAttrSet("seerr_notification_pushover.test", "id"),
				),
			},
			{
				Config: testAccNotificationPushoverConfig("app-token-update", "user-token-update", "siren"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_notification_pushover.test", "pushover.sound", "siren"),
					resource.TestCheckResourceAttr("seerr_notification_pushover.test", "on_issue_created", "true"),
				),
			},
			{
				ResourceName:            "seerr_notification_pushover.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"pushover.access_token", "pushover.user_token"},
			},
			{
				Config: testAccNotificationPushoverConfig("app-token-update", "user-token-update", "siren") + testAccNotificationPushoverDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_notification_pushover.current", "enabled", "true"),
					resource.TestCheckResourceAttr("data.seerr_notification_pushover.current", "pushover.sound", "siren"),
					resource.TestCheckResourceAttr("data.seerr_notification_pushover.current", "on_issue_created", "true"),
				),
			},
		},
	})
}

func testAccNotificationNtfyConfig(url, topic, token string, priority int, embedPoster bool) string {
	return fmt.Sprintf(`
resource "seerr_notification_ntfy" "test" {
  enabled      = true
  embed_poster = %[5]t

  ntfy = {
    url               = %[1]q
    topic             = %[2]q
    auth_method_token = true
    token             = %[3]q
    priority          = %[4]d
  }

  notification_types = ["MEDIA_PENDING", "ISSUE_CREATED"]
}
`, url, topic, token, priority, embedPoster)
}

func testAccNotificationNtfyDataSourceConfig() string {
	return `
data "seerr_notification_ntfy" "current" {}
`
}

func testAccNotificationPushoverConfig(accessToken, userToken, sound string) string {
	return fmt.Sprintf(`
resource "seerr_notification_pushover" "test" {
  enabled = true

  pushover = {
    access_token = %[1]q
    user_token   = %[2]q
    sound        = %[3]q
  }

  notification_types = ["MEDIA_PENDING", "ISSUE_CREATED"]
}
`, accessToken, userToken, sound)
}

func testAccNotificationPushoverDataSourceConfig() string {
	return `
data "seerr_notification_pushover" "current" {}
`
}
