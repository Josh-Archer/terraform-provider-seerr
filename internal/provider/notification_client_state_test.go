package provider

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type fakeNotificationAttributeReader struct {
	values    map[string]any
	requested []string
}

func (r *fakeNotificationAttributeReader) GetAttribute(_ context.Context, p path.Path, target any) diag.Diagnostics {
	key := p.String()
	r.requested = append(r.requested, key)

	value, ok := r.values[key]
	if !ok {
		return nil
	}

	reflect.ValueOf(target).Elem().Set(reflect.ValueOf(value))
	return nil
}

type fakeNotificationAttributeWriter struct {
	writes map[string]any
}

func (w *fakeNotificationAttributeWriter) SetAttribute(_ context.Context, p path.Path, value any) diag.Diagnostics {
	if w.writes == nil {
		w.writes = map[string]any{}
	}
	w.writes[p.String()] = value
	return nil
}

func TestReadNotificationClientModelReadsOnlyActiveAttribute(t *testing.T) {
	t.Parallel()

	reader := &fakeNotificationAttributeReader{
		values: map[string]any{
			path.Root("enabled").String():            types.BoolValue(true),
			path.Root("embed_poster").String():       types.BoolValue(true),
			path.Root("notification_types").String(): types.SetValueMust(types.StringType, []attr.Value{types.StringValue("MEDIA_PENDING"), types.StringValue("ISSUE_CREATED")}),
			path.Root("ntfy").String(): &NotificationAgentNtfyModel{
				Url:      types.StringValue("https://ntfy.example.com"),
				Topic:    types.StringValue("terraform"),
				Priority: types.Int64Value(4),
			},
			path.Root("pushover").String(): &NotificationAgentPushoverModel{
				Sound: types.StringValue("bike"),
			},
		},
	}

	data, diags := readNotificationClientModel(context.Background(), reader, "ntfy")
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if data.Ntfy == nil || data.Ntfy.Topic.ValueString() != "terraform" {
		t.Fatalf("expected ntfy payload to be populated, got %#v", data.Ntfy)
	}
	if data.Pushover != nil {
		t.Fatalf("expected pushover payload to remain nil, got %#v", data.Pushover)
	}

	assertPathRequested(t, reader.requested, path.Root("ntfy"))
	assertPathNotRequested(t, reader.requested, path.Root("pushover"))
}

func TestSetNotificationClientStateWritesOnlyActiveAttribute(t *testing.T) {
	t.Parallel()

	writer := &fakeNotificationAttributeWriter{}
	data := &NotificationAgentModel{
		ID:                types.StringValue("ntfy"),
		Agent:             types.StringValue("ntfy"),
		Enabled:           types.BoolValue(true),
		EmbedPoster:       types.BoolValue(false),
		NotificationTypes: types.SetValueMust(types.StringType, []attr.Value{types.StringValue("MEDIA_PENDING"), types.StringValue("ISSUE_CREATED")}),
		Ntfy:              &NotificationAgentNtfyModel{Topic: types.StringValue("terraform")},
		Pushover:          &NotificationAgentPushoverModel{Sound: types.StringValue("bike")},
	}

	diags := setNotificationClientState(context.Background(), writer, data)
	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	if _, ok := writer.writes[path.Root("ntfy").String()]; !ok {
		t.Fatalf("expected ntfy attribute to be written")
	}
	if _, ok := writer.writes[path.Root("pushover").String()]; ok {
		t.Fatalf("did not expect pushover attribute to be written")
	}

	gotEnabled, ok := writer.writes[path.Root("enabled").String()].(types.Bool)
	if !ok || !gotEnabled.ValueBool() {
		t.Fatalf("expected enabled=true write, got %#v", writer.writes[path.Root("enabled").String()])
	}
}

func assertPathRequested(t *testing.T, requested []string, expected path.Path) {
	t.Helper()
	expectedKey := expected.String()
	for _, key := range requested {
		if key == expectedKey {
			return
		}
	}
	t.Fatalf("expected path %q to be requested, got %v", expectedKey, requested)
}

func assertPathNotRequested(t *testing.T, requested []string, unexpected path.Path) {
	t.Helper()
	unexpectedKey := unexpected.String()
	for _, key := range requested {
		if key == unexpectedKey {
			t.Fatalf("did not expect path %q to be requested, got %v", unexpectedKey, requested)
		}
	}
}

func TestBuildPayloadNotificationOptionHygiene(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Case 1: Priority is Unknown (omitted in plan for Optional+Computed attribute)
	modelUnknown := &NotificationAgentModel{
		Agent:       types.StringValue("ntfy"),
		Enabled:     types.BoolValue(true),
		EmbedPoster: types.BoolValue(false),
		Ntfy: &NotificationAgentNtfyModel{
			Url:      types.StringValue("https://ntfy.example.com"),
			Topic:    types.StringValue("test"),
			Priority: types.Int64Unknown(),
		},
	}
	payloadStr, err := buildPayload(ctx, modelUnknown)
	if err != nil {
		t.Fatalf("buildPayload failed: %v", err)
	}
	if reflect.ValueOf(payloadStr).Kind() == reflect.String && (reflect.ValueOf(payloadStr).Len() == 0 || (payloadStr != "" && (func() bool {
		var m struct {
			Options map[string]interface{} `json:"options"`
		}
		_ = json.Unmarshal([]byte(payloadStr), &m)
		_, exists := m.Options["priority"]
		return exists
	})())) {
		t.Fatalf("expected priority key to be omitted from payload options when unknown, got %s", payloadStr)
	}

	// Case 2: Priority is Null
	modelNull := &NotificationAgentModel{
		Agent:       types.StringValue("ntfy"),
		Enabled:     types.BoolValue(true),
		EmbedPoster: types.BoolValue(false),
		Ntfy: &NotificationAgentNtfyModel{
			Url:      types.StringValue("https://ntfy.example.com"),
			Topic:    types.StringValue("test"),
			Priority: types.Int64Null(),
		},
	}
	payloadStrNull, err := buildPayload(ctx, modelNull)
	if err != nil {
		t.Fatalf("buildPayload failed: %v", err)
	}
	var mNull struct {
		Options map[string]interface{} `json:"options"`
	}
	if err := json.Unmarshal([]byte(payloadStrNull), &mNull); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	if _, exists := mNull.Options["priority"]; exists {
		t.Fatalf("expected priority key to be omitted from payload options when null, got %s", payloadStrNull)
	}

	// Case 3: Priority is explicitly set (e.g. 5)
	modelSet := &NotificationAgentModel{
		Agent:       types.StringValue("ntfy"),
		Enabled:     types.BoolValue(true),
		EmbedPoster: types.BoolValue(false),
		Ntfy: &NotificationAgentNtfyModel{
			Url:      types.StringValue("https://ntfy.example.com"),
			Topic:    types.StringValue("test"),
			Priority: types.Int64Value(5),
		},
	}
	payloadStrSet, err := buildPayload(ctx, modelSet)
	if err != nil {
		t.Fatalf("buildPayload failed: %v", err)
	}
	var mSet struct {
		Options map[string]interface{} `json:"options"`
	}
	if err := json.Unmarshal([]byte(payloadStrSet), &mSet); err != nil {
		t.Fatalf("json unmarshal failed: %v", err)
	}
	pVal, exists := mSet.Options["priority"]
	if !exists || pVal != float64(5) {
		t.Fatalf("expected priority key to be 5 in payload options, got %#v in %s", pVal, payloadStrSet)
	}
}

func TestNotificationAgentSchemasOptionalComputedAudit(t *testing.T) {
	t.Parallel()

	options := notificationAgentResourceOptionAttributes()
	for agentName, agentAttr := range options {
		nested, ok := agentAttr.(schema.SingleNestedAttribute)
		if !ok {
			continue
		}
		for attrName, attr := range nested.Attributes {
			isRequired := attr.IsRequired()
			isComputed := attr.IsComputed()
			isOptional := attr.IsOptional()

			if !isRequired {
				if !isOptional {
					t.Errorf("agent %s attribute %s: expected Optional to be true", agentName, attrName)
				}
				if !isComputed {
					t.Errorf("agent %s attribute %s: expected Computed to be true for state hygiene", agentName, attrName)
				}
			}
		}
	}
}
