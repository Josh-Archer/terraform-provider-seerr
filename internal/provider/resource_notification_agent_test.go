package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAccNotificationAgentResource(t *testing.T) {
	// Skip acceptance test in this environment as it requires a Seerr instance
}

func TestNotificationBitmaskMapping(t *testing.T) {
	resource := &notificationAgentResource{}

	// Test case: request_pending (2) and issue_created (256)
	data := &NotificationAgentModel{
		OnRequestPending: types.BoolValue(true),
		OnIssueCreated:   types.BoolValue(true),
		TypesMask:        types.Int64Value(0),
	}

	payload := resource.buildPayload(data)
	expected := int64(2 | 256)
	if payload.Types != expected {
		t.Errorf("Expected mask %d, got %d", expected, payload.Types)
	}

	// Test case: parse back
	newData := &NotificationAgentModel{}
	resource.parsePayload(payload, newData)

	if !newData.OnRequestPending.ValueBool() {
		t.Error("OnRequestPending should be true after parsing")
	}
	if !newData.OnIssueCreated.ValueBool() {
		t.Error("OnIssueCreated should be true after parsing")
	}
	if newData.OnMediaAvailable.ValueBool() {
		t.Error("OnMediaAvailable should be false")
	}

	// Test case: media_auto_requested (4096)
	data2 := &NotificationAgentModel{
		OnMediaAutoRequested: types.BoolValue(true),
		TypesMask:            types.Int64Value(0),
	}
	payload2 := resource.buildPayload(data2)
	if payload2.Types != 4096 {
		t.Errorf("Expected mask 4096, got %d", payload2.Types)
	}

	newData2 := &NotificationAgentModel{}
	resource.parsePayload(payload2, newData2)
	if !newData2.OnMediaAutoRequested.ValueBool() {
		t.Error("OnMediaAutoRequested should be true after parsing")
	}
}
