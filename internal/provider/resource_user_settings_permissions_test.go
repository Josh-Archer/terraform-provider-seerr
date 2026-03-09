package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserSettingsPermissionsResource(t *testing.T) {
	username := "test_settings_user"
	email := "test_settings@example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Setup: Create User
			{
				Config: testAccUserSettingsPermissionsConfig_Base(username, email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("seerr_user.test", "id"),
				),
			},
			// Test: Create Settings Permissions
			{
				Config: testAccUserSettingsPermissionsConfig_Base(username, email) + testAccUserSettingsPermissionsConfig_Step1(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user_settings_permissions.test", "auto_approve_movies", "true"),
					resource.TestCheckResourceAttr("seerr_user_settings_permissions.test", "auto_approve_tv", "false"),
				),
			},
			// Test: Update
			{
				Config: testAccUserSettingsPermissionsConfig_Base(username, email) + testAccUserSettingsPermissionsConfig_Step2(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user_settings_permissions.test", "auto_approve_movies", "false"),
					resource.TestCheckResourceAttr("seerr_user_settings_permissions.test", "auto_approve_tv", "true"),
				),
			},
			// Test: Import
			{
				ResourceName:      "seerr_user_settings_permissions.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserSettingsPermissionsConfig_Base(username, email string) string {
	return fmt.Sprintf(`
resource "seerr_user" "test" {
  username    = %[1]q
  email       = %[2]q
  permissions = 0
}
`, username, email)
}

func testAccUserSettingsPermissionsConfig_Step1() string {
	return `
resource "seerr_user_settings_permissions" "test" {
  user_id             = seerr_user.test.id
  auto_approve_movies = true
  auto_approve_tv     = false
}
`
}

func testAccUserSettingsPermissionsConfig_Step2() string {
	return `
resource "seerr_user_settings_permissions" "test" {
  user_id             = seerr_user.test.id
  auto_approve_movies = false
  auto_approve_tv     = true
}
`
}
