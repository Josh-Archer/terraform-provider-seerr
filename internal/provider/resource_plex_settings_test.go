package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPlexSettingsResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccPlexSettingsResourceConfig("127.0.0.1", 32400, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_plex_settings.test", "ip", "127.0.0.1"),
					resource.TestCheckResourceAttr("seerr_plex_settings.test", "port", "32400"),
					resource.TestCheckResourceAttr("seerr_plex_settings.test", "use_ssl", "false"),
					resource.TestCheckResourceAttrSet("seerr_plex_settings.test", "name"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "seerr_plex_settings.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"response_json", // API responses may vary slightly in formatting
				},
			},
			// Update and Read testing
			{
				Config: testAccPlexSettingsResourceConfig("localhost", 32401, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_plex_settings.test", "ip", "localhost"),
					resource.TestCheckResourceAttr("seerr_plex_settings.test", "port", "32401"),
					resource.TestCheckResourceAttr("seerr_plex_settings.test", "use_ssl", "true"),
				),
			},
		},
	})
}

func testAccPlexSettingsResourceConfig(ip string, port int, useSSL bool) string {
	return fmt.Sprintf(`
resource "seerr_plex_settings" "test" {
  ip      = %[1]q
  port    = %[2]d
  use_ssl = %[3]t
}
`, ip, port, useSSL)
}
