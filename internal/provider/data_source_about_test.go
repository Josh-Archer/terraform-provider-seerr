package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAboutDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAboutDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_about.test", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_about.test", "version"),
					resource.TestCheckResourceAttrSet("data.seerr_about.test", "total_requests"),
					resource.TestCheckResourceAttrSet("data.seerr_about.test", "total_media_items"),
					resource.TestCheckResourceAttrSet("data.seerr_about.test", "app_data_path"),
				),
			},
		},
	})
}

func testAccAboutDataSourceConfig() string {
	return providerConfig + `
data "seerr_about" "test" {}
`
}
