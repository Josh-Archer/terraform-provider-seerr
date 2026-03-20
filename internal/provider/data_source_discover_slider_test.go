package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDiscoverSliderDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "seerr_discover_slider" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_discover_slider.test", "id", "settings"),
					resource.TestCheckResourceAttrSet("data.seerr_discover_slider.test", "sliders.#"),
				),
			},
		},
	})
}
