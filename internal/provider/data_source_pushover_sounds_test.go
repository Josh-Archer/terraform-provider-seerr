package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccPushoverSoundsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPushoverSoundsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_pushover_sounds.test", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_pushover_sounds.test", "sounds.#"),
				),
			},
		},
	})
}

func testAccPushoverSoundsDataSourceConfig() string {
	return providerConfig + `
data "seerr_pushover_sounds" "test" {
  token = "test-token"
}
`
}
