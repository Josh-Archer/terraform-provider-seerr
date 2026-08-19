package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRequestsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "seerr_requests" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_requests.all", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_requests.all", "requests.#"),
				),
			},
			{
				Config: `
data "seerr_requests" "filtered" {
  filter_status = 1
  filter_media_type = "movie"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_requests.filtered", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_requests.filtered", "requests.#"),
					resource.TestCheckResourceAttrSet("data.seerr_requests.filtered", "total"),
				),
			},
		},
	})
}
