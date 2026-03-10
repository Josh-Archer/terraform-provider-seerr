package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccJobsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read all jobs, there are always default jobs configured in Seerr
			{
				Config: `
data "seerr_jobs" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_jobs.all", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_jobs.all", "jobs.#"),
				),
			},
		},
	})
}
