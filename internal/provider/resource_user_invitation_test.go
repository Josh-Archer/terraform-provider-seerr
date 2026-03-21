package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("seerr_user_invitation", &resource.Sweeper{
		Name: "seerr_user_invitation",
		F:    sweepUserInvitation,
	})
}

func sweepUserInvitation(region string) error {
	client, err := testAccClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	res, err := client.Request(ctx, "GET", "/api/v1/user/invites", "", nil)
	if err != nil {
		return fmt.Errorf("error fetching invitations: %s", err)
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("error fetching invitations: status %d", res.StatusCode)
	}

	var results []map[string]any
	if err := json.Unmarshal(res.Body, &results); err != nil {
		return fmt.Errorf("error parsing invitations response: %s", err)
	}

	for _, inv := range results {
		idRaw, ok := inv["id"]
		if !ok {
			continue
		}

		var id string
		switch v := idRaw.(type) {
		case float64:
			id = fmt.Sprintf("%.0f", v)
		case string:
			id = v
		}

		log.Printf("[INFO][SWEEPER] Deleting invitation %s", id)
		delRes, err := client.Request(ctx, "DELETE", "/api/v1/user/invite/"+id, "", nil)
		if err != nil {
			log.Printf("[ERROR][SWEEPER] Error deleting invitation %s: %s", id, err)
			continue
		}
		if !StatusIsOK(delRes.StatusCode) {
			log.Printf("[ERROR][SWEEPER] Error deleting invitation %s: status %d", id, delRes.StatusCode)
		}
	}

	return nil
}

func TestAccUserInvitationResource(t *testing.T) {
	email := "test_invite@example.com"

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Verify
			{
				Config: testAccUserInvitationResourceConfig(email),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("seerr_user_invitation.test", "email", email),
					resource.TestCheckResourceAttrSet("seerr_user_invitation.test", "id"),
				),
			},
			// ImportState
			{
				ResourceName:      "seerr_user_invitation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserInvitationResourceConfig(email string) string {
	return fmt.Sprintf(`
resource "seerr_user_invitation" "test" {
  email = %[1]q
}

data "seerr_user_invitations" "all" {
  depends_on = [seerr_user_invitation.test]
}
`, email)
}
