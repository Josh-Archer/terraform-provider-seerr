package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIssuesDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read all issues
			{
				Config: `
data "seerr_issues" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_issues.all", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_issues.all", "issues.#"),
				),
			},
		},
	})
}
