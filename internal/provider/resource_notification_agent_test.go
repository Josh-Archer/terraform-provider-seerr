package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationAgentResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { /* test API configuration */ },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
				provider "seerr" {
					url     = "http://localhost:5055"
					api_key = "test-api-key"
				}

				resource "seerr_notification_agent" "test_discord" {
					agent   = "discord"
					enabled = true
					types   = 2

					discord = {
						webhook_url     = "https://discord.com/api/webhooks/test"
						bot_username    = "SeerrBot"
						enable_mentions = true
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_notification_agent.test_discord", "agent", "discord"),
					resource.TestCheckResourceAttr("seerr_notification_agent.test_discord", "enabled", "true"),
					resource.TestCheckResourceAttr("seerr_notification_agent.test_discord", "types", "2"),
					resource.TestCheckResourceAttr("seerr_notification_agent.test_discord", "discord.bot_username", "SeerrBot"),
					resource.TestCheckResourceAttr("seerr_notification_agent.test_discord", "discord.webhook_url", "https://discord.com/api/webhooks/test"),
				),
			},
			{
				Config: `
				provider "seerr" {
					url     = "http://localhost:5055"
					api_key = "test-api-key"
				}

				resource "seerr_notification_agent" "test_telegram" {
					agent   = "telegram"
					enabled = false
					types   = 4

					telegram = {
						bot_api = "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
						chat_id = "-1001234567890"
					}
				}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_notification_agent.test_telegram", "agent", "telegram"),
					resource.TestCheckResourceAttr("seerr_notification_agent.test_telegram", "enabled", "false"),
					resource.TestCheckResourceAttr("seerr_notification_agent.test_telegram", "types", "4"),
					resource.TestCheckResourceAttr("seerr_notification_agent.test_telegram", "telegram.chat_id", "-1001234567890"),
					resource.TestMatchResourceAttr("seerr_notification_agent.test_telegram", "telegram.bot_api", regexp.MustCompile(`^123456`)),
				),
			},
		},
	})
}
