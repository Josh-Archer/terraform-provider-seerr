package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAccNotificationAgentResource(t *testing.T) {
	// Skip acceptance test in this environment as it requires a Seerr instance
}

func TestNotificationBitmaskMapping(t *testing.T) {
	// Test case: request_pending (2) and issue_created (256)
	data := &NotificationAgentModel{
		Agent:            types.StringValue("discord"),
		OnRequestPending: types.BoolValue(true),
		OnIssueCreated:   types.BoolValue(true),
		TypesMask:        types.Int64Value(0),
	}

	payloadStr, err := buildPayload(data)
	if err != nil {
		t.Fatalf("buildPayload failed: %v", err)
	}
	payload := &notificationAgentPayload{}
	if err := json.Unmarshal([]byte(payloadStr), payload); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	expected := int64(2 | 256)
	if payload.Types != expected {
		t.Errorf("Expected mask %d, got %d", expected, payload.Types)
	}

	// Test case: parse back
	newData := &NotificationAgentModel{Agent: types.StringValue("discord")}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	if err := parsePayload(newData, payloadBytes); err != nil {
		t.Fatalf("parsePayload failed: %v", err)
	}

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
		Agent:                types.StringValue("discord"),
		OnMediaAutoRequested: types.BoolValue(true),
		TypesMask:            types.Int64Value(0),
	}
	payloadStr2, err := buildPayload(data2)
	if err != nil {
		t.Fatalf("buildPayload failed: %v", err)
	}
	payload2 := &notificationAgentPayload{}
	if err := json.Unmarshal([]byte(payloadStr2), payload2); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if payload2.Types != 4096 {
		t.Errorf("Expected mask 4096, got %d", payload2.Types)
	}

	newData2 := &NotificationAgentModel{Agent: types.StringValue("discord")}
	payloadBytes2, err := json.Marshal(payload2)
	if err != nil {
		t.Fatalf("json marshal failed: %v", err)
	}
	if err := parsePayload(newData2, payloadBytes2); err != nil {
		t.Fatalf("parsePayload failed: %v", err)
	}
	if !newData2.OnMediaAutoRequested.ValueBool() {
		t.Error("OnMediaAutoRequested should be true after parsing")
	}
}

func TestNotificationAgentUnknownResolution(t *testing.T) {
	// This test simulates the case where a field is Unknown in the plan
	// and verifies that parsePayload sets it to a Known value from the API response.
	data := &NotificationAgentModel{
		Agent:           types.StringValue("pushover"),
		OnMediaFollowed: types.BoolUnknown(), // Simulate Unknown from Plan
	}

	// Mock API response body
	responseBody := `{"enabled":true,"types":2048,"options":{"accessToken":"secret","userToken":"secret"}}`

	if err := parsePayload(data, []byte(responseBody)); err != nil {
		t.Fatalf("parsePayload failed: %v", err)
	}

	if data.OnMediaFollowed.IsUnknown() {
		t.Error("OnMediaFollowed should NOT be Unknown after parsePayload")
	}
	if !data.OnMediaFollowed.ValueBool() {
		t.Error("OnMediaFollowed should be true (mask 2048)")
	}
}
