package provider

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type notificationStateWriter struct {
	paths []string
}

func (w *notificationStateWriter) SetAttribute(_ context.Context, p path.Path, _ any) diag.Diagnostics {
	w.paths = append(w.paths, p.String())
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

func TestSetNotificationClientStateDoesNotWriteLegacyTypes(t *testing.T) {
	ctx := context.Background()
	writer := &notificationStateWriter{}
	data := &NotificationAgentModel{
		Agent:             types.StringValue("ntfy"),
		NotificationTypes: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("ISSUE_CREATED")}),
	}

	if diags := setNotificationClientState(ctx, writer, data); diags.HasError() {
		t.Fatalf("setNotificationClientState returned diagnostics: %v", diags)
	}

	for _, p := range writer.paths {
		if p == "types" {
			t.Fatalf("setNotificationClientState wrote legacy path %q", p)
		}
	}
}
