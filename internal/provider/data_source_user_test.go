package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestPopulateUserDataSourceModelPreservesAPICasing(t *testing.T) {
	data := UserDataSourceModel{}
	err := populateUserDataSourceModel(&data, map[string]any{
		"id":          float64(42),
		"email":       "Case.Sensitive@Example.COM",
		"username":    "MixedCaseUser",
		"permissions": "7",
	})
	if err != nil {
		t.Fatalf("populate user data source model: %v", err)
	}

	if got, want := data.ID, types.StringValue("42"); !got.Equal(want) {
		t.Fatalf("id = %q, want %q", got.ValueString(), want.ValueString())
	}
	if got, want := data.Email, types.StringValue("Case.Sensitive@Example.COM"); !got.Equal(want) {
		t.Fatalf("email = %q, want %q", got.ValueString(), want.ValueString())
	}
	if got, want := data.Username, types.StringValue("MixedCaseUser"); !got.Equal(want) {
		t.Fatalf("username = %q, want %q", got.ValueString(), want.ValueString())
	}
	if got := data.Permissions.ValueInt64(); got != 7 {
		t.Fatalf("permissions = %d, want 7", got)
	}
}

func TestAccUserDataSource(t *testing.T) {
	username := "terraform_ds_test_user"
	email := "ds_test_user@example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Setup: Create a user to look up
			{
				Config: testAccUserResourceConfig(username, email, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("seerr_user.test", "id"),
				),
			},
			// Test: Look up by email
			{
				Config: testAccUserResourceConfig(username, email, 0) + testAccUserDataSourceConfigByEmail(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_user.by_email", "username", username),
					resource.TestCheckResourceAttr("data.seerr_user.by_email", "permissions", "0"),
					resource.TestCheckResourceAttrSet("data.seerr_user.by_email", "id"),
				),
			},
			// Test: Look up by username
			{
				Config: testAccUserResourceConfig(username, email, 0) + testAccUserDataSourceConfigByUsername(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_user.by_username", "email", email),
					resource.TestCheckResourceAttr("data.seerr_user.by_username", "permissions", "0"),
					resource.TestCheckResourceAttrSet("data.seerr_user.by_username", "id"),
				),
			},
			// Test: Look up by email with different case
			{
				Config: testAccUserResourceConfig(username, email, 0) + testAccUserDataSourceConfigByEmail("DS_TEST_USER@EXAMPLE.COM"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_user.by_email", "email", email), // Should be normalized to lowercase
					resource.TestCheckResourceAttr("data.seerr_user.by_email", "username", username),
					resource.TestCheckResourceAttrSet("data.seerr_user.by_email", "id"),
				),
			},
		},
	})
}

func testAccUserDataSourceConfigByEmail(email string) string {
	return fmt.Sprintf(`
data "seerr_user" "by_email" {
  email = %[1]q
}
`, email)
}

func testAccUserDataSourceConfigByUsername(username string) string {
	return fmt.Sprintf(`
data "seerr_user" "by_username" {
  username = %[1]q
}
`, username)
}
