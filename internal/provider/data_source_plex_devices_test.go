package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPlexDevicesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPlexDevicesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_plex_devices.test", "id", "plex_devices"),
				),
			},
		},
	})
}

func testAccPlexDevicesDataSourceConfig() string {
	return providerConfig + `
data "seerr_plex_devices" "test" {}
`
}
