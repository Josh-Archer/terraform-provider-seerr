package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestNotificationAgentsDataSourceReadsPerAgentEndpoints(t *testing.T) {
	rootCalled := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/settings/notifications":
			rootCalled = true
			http.NotFound(w, r)
		case "/api/v1/settings/notifications/discord":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"enabled":true,"embedPoster":false}`))
		case "/api/v1/settings/notifications/ntfy":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"enabled":false,"embedPoster":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	d := &NotificationAgentsDataSource{
		client: NewClient(baseURL, "abc123", "test-agent", false, defaultRequestTimeout),
	}
	agents, err := d.readNotificationAgentSummaries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rootCalled {
		t.Fatal("aggregate notification endpoint should not be requested")
	}
	if got, want := len(agents), 2; got != want {
		t.Fatalf("expected %d agents, got %d", want, got)
	}
	if got := agents[0].Agent.ValueString(); got != "discord" {
		t.Fatalf("expected first agent discord, got %q", got)
	}
	if got := agents[0].Enabled.ValueBool(); !got {
		t.Fatalf("expected discord enabled true, got %v", got)
	}
	if got := agents[1].Agent.ValueString(); got != "ntfy" {
		t.Fatalf("expected second agent ntfy, got %q", got)
	}
	if got := agents[1].EmbedPoster.ValueBool(); !got {
		t.Fatalf("expected ntfy embed_poster true, got %v", got)
	}
}

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
