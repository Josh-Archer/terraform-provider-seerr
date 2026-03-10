package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNotificationAgentsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read all agents. Seerr returns the list of all supported agents whether enabled or not.
			{
				Config: `
data "seerr_notification_agents" "all" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.seerr_notification_agents.all", "id"),
					resource.TestCheckResourceAttrSet("data.seerr_notification_agents.all", "agents.#"),
				),
			},
		},
	})
}
