package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
