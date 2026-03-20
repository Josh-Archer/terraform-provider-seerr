package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCurrentUserDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCurrentUserDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_current_user.test", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_current_user.test", "username"),
					resource.TestCheckResourceAttrSet("data.seerr_current_user.test", "email"),
					resource.TestCheckResourceAttrSet("data.seerr_current_user.test", "permissions"),
				),
			},
		},
	})
}

func testAccCurrentUserDataSourceConfig() string {
	return providerConfig + `
data "seerr_current_user" "test" {}
`
}
