package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func init() {
	resource.AddTestSweepers("seerr_notification_agent", &resource.Sweeper{
		Name: "seerr_notification_agent",
		F:    sweepNotificationAgent,
	})
}

func sweepNotificationAgent(region string) error {
	client, err := testAccClient()
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	ctx := context.Background()
	res, err := client.Request(ctx, "GET", "/api/v1/settings/notifications", "", nil)
	if err != nil {
		return fmt.Errorf("error fetching notification agents: %s", err)
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("error fetching notification agents: status %d", res.StatusCode)
	}

	var settings map[string]any
	if err := json.Unmarshal(res.Body, &settings); err != nil {
		return fmt.Errorf("error parsing notification agents: %s", err)
	}

	disablePayload := `{"enabled":false,"types":0,"options":{}}`

	for agentName, agentData := range settings {
		agentMap, ok := agentData.(map[string]any)
		if !ok {
			continue
		}

		enabled, _ := agentMap["enabled"].(bool)
		if !enabled {
			continue
		}

		log.Printf("[INFO][SWEEPER] Disabling notification agent %s", agentName)
		delRes, err := client.Request(ctx, "POST", "/api/v1/settings/notifications/"+agentName, disablePayload, nil)
		if err != nil {
			log.Printf("[ERROR][SWEEPER] Error disabling agent %s: %s", agentName, err)
			continue
		}
		if !StatusIsOK(delRes.StatusCode) {
			log.Printf("[ERROR][SWEEPER] Error disabling agent %s: status %d", agentName, delRes.StatusCode)
		}
	}
	return nil
}

func TestAccNotificationAgentResource(t *testing.T) {
	// Skip acceptance test in this environment as it requires a Seerr instance
}

func TestNotificationBitmaskMapping(t *testing.T) {
	ctx := context.Background()
	// Test case: request_pending (2) and issue_created (256)
	data := &NotificationAgentModel{
		Agent:             types.StringValue("discord"),
		NotificationTypes: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("MEDIA_PENDING"), types.StringValue("ISSUE_CREATED")}),
	}

	payloadStr, err := buildPayload(ctx, data)
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
	if err := parsePayload(ctx, newData, payloadBytes); err != nil {
		t.Fatalf("parsePayload failed: %v", err)
	}

	var parsedTypes []string
	newData.NotificationTypes.ElementsAs(ctx, &parsedTypes, false)
	foundPending := false
	foundIssue := false
	for _, pt := range parsedTypes {
		if pt == "MEDIA_PENDING" {
			foundPending = true
		}
		if pt == "ISSUE_CREATED" {
			foundIssue = true
		}
	}
	if !foundPending {
		t.Error("MEDIA_PENDING should be true after parsing")
	}
	if !foundIssue {
		t.Error("ISSUE_CREATED should be true after parsing")
	}

	// Test case: media_auto_requested (4096)
	data2 := &NotificationAgentModel{
		Agent:             types.StringValue("discord"),
		NotificationTypes: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("MEDIA_AUTO_REQUESTED")}),
	}
	payloadStr2, err := buildPayload(ctx, data2)
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
	if err := parsePayload(ctx, newData2, payloadBytes2); err != nil {
		t.Fatalf("parsePayload failed: %v", err)
	}
	var parsedTypes2 []string
	newData2.NotificationTypes.ElementsAs(ctx, &parsedTypes2, false)
	foundAuto := false
	for _, pt := range parsedTypes2 {
		if pt == "MEDIA_AUTO_REQUESTED" {
			foundAuto = true
		}
	}
	if !foundAuto {
		t.Error("MEDIA_AUTO_REQUESTED should be true after parsing")
	}
}

func TestNotificationAgentUnknownResolution(t *testing.T) {
	ctx := context.Background()
	// This test simulates the case where a field is Unknown in the plan
	// and verifies that parsePayload sets it to a Known value from the API response.
	data := &NotificationAgentModel{
		Agent:             types.StringValue("pushover"),
		NotificationTypes: types.SetUnknown(types.StringType), // Simulate Unknown from Plan
	}

	// Mock API response body
	responseBody := `{"enabled":true,"types":2048,"options":{"accessToken":"secret","userToken":"secret"}}`

	if err := parsePayload(ctx, data, []byte(responseBody)); err != nil {
		t.Fatalf("parsePayload failed: %v", err)
	}

	if data.NotificationTypes.IsUnknown() {
		t.Error("NotificationTypes should NOT be Unknown after parsePayload")
	}
	var parsedTypes []string
	data.NotificationTypes.ElementsAs(ctx, &parsedTypes, false)
	found := false
	for _, pt := range parsedTypes {
		if pt == "ISSUE_REOPENED" {
			found = true
		}
	}
	if !found {
		t.Error("ISSUE_REOPENED should be true (mask 2048)")
	}
}

func TestNotificationTypesMaskIgnoresMissingDefaultAndDerivesFromFlags(t *testing.T) {
	ctx := context.Background()
	data := &NotificationAgentModel{
		NotificationTypes: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("MEDIA_PENDING"), types.StringValue("ISSUE_CREATED")}),
	}

	if got, want := notificationTypesMask(ctx, data), int64(258); got != want {
		t.Fatalf("notificationTypesMask() = %d, want %d", got, want)
	}
}
