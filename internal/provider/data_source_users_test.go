package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUsersDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read all users, there should be at least one (the admin who created the token or the imported test user)
			{
				Config: `
data "seerr_users" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_users.all", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_users.all", "users.#"),
				),
			},
		},
	})
}
