package provider

import (
	"testing"

	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// Webhook URLs embed secrets (Discord/Slack tokens), so every webhook_url
// attribute must be Sensitive in both resource and data source schemas.
func TestNotificationAgentWebhookURLSensitive(t *testing.T) {
	found := 0
	for agent, attr := range notificationAgentResourceOptionAttributes() {
		nested, ok := attr.(schema.SingleNestedAttribute)
		if !ok {
			continue
		}
		if w, ok := nested.Attributes["webhook_url"]; ok {
			found++
			if !w.IsSensitive() {
				t.Errorf("resource %s.webhook_url is not Sensitive", agent)
			}
		}
	}
	if found != 4 {
		t.Errorf("expected 4 resource webhook_url attributes, found %d", found)
	}

	found = 0
	for agent, attr := range notificationAgentDataSourceOptionAttributes() {
		nested, ok := attr.(dschema.SingleNestedAttribute)
		if !ok {
			continue
		}
		if w, ok := nested.Attributes["webhook_url"]; ok {
			found++
			if !w.IsSensitive() {
				t.Errorf("data source %s.webhook_url is not Sensitive", agent)
			}
		}
	}
	if found != 4 {
		t.Errorf("expected 4 data source webhook_url attributes, found %d", found)
	}
}
