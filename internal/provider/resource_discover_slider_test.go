package provider

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResolveSliderIDsReusesExistingIDs(t *testing.T) {
	current := []DiscoverSliderItemModel{
		{
			ID:        types.Int64Value(101),
			Type:      types.Int64Value(1),
			IsBuiltIn: types.BoolValue(true),
			Enabled:   types.BoolValue(true),
		},
		{
			ID:      types.Int64Value(202),
			Type:    types.Int64Value(9),
			Enabled: types.BoolValue(true),
			Title:   types.StringValue("Custom"),
			Data:    types.StringValue("trending"),
		},
	}

	desired := []DiscoverSliderItemModel{
		{
			Type:    types.Int64Value(1),
			Enabled: types.BoolValue(false),
		},
		{
			Type:    types.Int64Value(9),
			Enabled: types.BoolValue(true),
			Title:   types.StringValue("Custom"),
			Data:    types.StringValue("trending"),
		},
	}

	resolved := resolveSliderIDs(current, desired)
	if got := resolved[0].ID.ValueInt64(); got != 101 {
		t.Fatalf("expected built-in slider id 101, got %d", got)
	}
	if got := resolved[1].ID.ValueInt64(); got != 202 {
		t.Fatalf("expected custom slider id 202, got %d", got)
	}
	if !resolved[0].IsBuiltIn.ValueBool() {
		t.Fatal("expected built-in marker to be preserved")
	}
}

func TestFilterManagedSlidersReturnsOnlyTrackedEntries(t *testing.T) {
	current := []DiscoverSliderItemModel{
		{ID: types.Int64Value(1), Type: types.Int64Value(1), Enabled: types.BoolValue(true)},
		{ID: types.Int64Value(2), Type: types.Int64Value(2), Enabled: types.BoolValue(true)},
		{ID: types.Int64Value(3), Type: types.Int64Value(4), Enabled: types.BoolValue(true)},
	}
	managed := []DiscoverSliderItemModel{
		{ID: types.Int64Value(2), Type: types.Int64Value(2)},
		{ID: types.Int64Value(3), Type: types.Int64Value(4)},
	}

	filtered := filterManagedSliders(current, managed)
	if len(filtered) != 2 {
		t.Fatalf("expected 2 managed sliders, got %d", len(filtered))
	}
	if got := filtered[0].ID.ValueInt64(); got != 2 {
		t.Fatalf("expected first managed slider id 2, got %d", got)
	}
	if got := filtered[1].ID.ValueInt64(); got != 3 {
		t.Fatalf("expected second managed slider id 3, got %d", got)
	}
}

func TestNotificationAgentMissingReturnsTrueFor404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`Unknown notification agent`))
	}))
	defer srv.Close()

	base, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	resource := &NotificationClientResource{
		agent:  "pushover",
		client: NewClient(base, "abc123", "test-agent", false, 5*time.Second),
	}

	if !resource.notificationAgentMissing(t.Context()) {
		t.Fatal("expected notificationAgentMissing to return true for missing agent")
	}
}
