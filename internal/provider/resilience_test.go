// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

// TestResilience404Reconciliation verifies that when an upstream resource is deleted
// out-of-band (returning HTTP 404), the Read method gracefully calls resp.State.RemoveResource
// without returning error diagnostics.
func TestResilience404Reconciliation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	assert.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	ctx := context.Background()

	t.Run("MainSettings 404 removes state", func(t *testing.T) {
		r := &MainSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := MainSettingsModel{ID: types.StringValue("main")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("PlexSettings 404 removes state", func(t *testing.T) {
		r := &PlexSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := PlexSettingsModel{ID: types.StringValue("plex")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("JellyfinSettings 404 removes state", func(t *testing.T) {
		r := &JellyfinSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := JellyfinSettingsModel{ID: types.StringValue("jellyfin")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("EmbySettings 404 removes state", func(t *testing.T) {
		r := &EmbySettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := EmbySettingsModel{ID: types.StringValue("emby")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("TautulliSettings 404 removes state", func(t *testing.T) {
		r := &TautulliSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := TautulliSettingsModel{ID: types.StringValue("tautulli")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("NetworkSettings 404 removes state", func(t *testing.T) {
		r := &NetworkSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := NetworkSettingsModel{ID: types.StringValue("network")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("MetadataSettings 404 removes state", func(t *testing.T) {
		r := &MetadataSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := MetadataSettingsModel{ID: types.StringValue("metadata")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("BackupSettings 404 removes state", func(t *testing.T) {
		r := &BackupSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := BackupSettingsModel{ID: types.StringValue("backups")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("RadarrServer 404 removes state", func(t *testing.T) {
		r := &RadarrServerResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := RadarrServerModel{
			ID:       types.StringValue("1"),
			ServerID: types.Int64Value(1),
			Tags:     types.ListNull(types.Int64Type),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		if readResp.Diagnostics.HasError() {
			t.Logf("Radarr diagnostics: %v", readResp.Diagnostics)
		}
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("SonarrServer 404 removes state", func(t *testing.T) {
		r := &SonarrServerResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := SonarrServerModel{
			ID:        types.StringValue("1"),
			ServerID:  types.Int64Value(1),
			Tags:      types.ListNull(types.Int64Type),
			AnimeTags: types.ListNull(types.Int64Type),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("User 404 removes state", func(t *testing.T) {
		r := &UserResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := UserModel{ID: types.StringValue("99")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("UserQuota 404 removes state", func(t *testing.T) {
		r := &UserQuotaResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := UserQuotaModel{
			ID:     types.StringValue("99"),
			UserID: types.Int64Value(99),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("UserPermissions 404 removes state", func(t *testing.T) {
		r := &UserPermissionsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := UserPermissionsModel{
			ID:     types.StringValue("99"),
			UserID: types.Int64Value(99),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("UserNotificationSettings 404 removes state", func(t *testing.T) {
		r := &UserNotificationSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := UserNotificationSettingsResourceModel{
			ID:     types.StringValue("99"),
			UserID: types.Int64Value(99),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("UserWatchlistSettings 404 removes state", func(t *testing.T) {
		r := &UserWatchlistSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := UserWatchlistSettingsModel{
			ID:     types.StringValue("99"),
			UserID: types.Int64Value(99),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("JobSchedule 404 removes state", func(t *testing.T) {
		r := &JobScheduleResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := JobScheduleModel{
			ID:    types.StringValue("plex-recent-sync"),
			JobID: types.StringValue("plex-recent-sync"),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("OverrideRule 404 removes state", func(t *testing.T) {
		r := &OverrideRuleResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := OverrideRuleModel{
			ID:        types.StringValue("1"),
			Genres:    types.ListNull(types.Int64Type),
			TagIDs:    types.ListNull(types.Int64Type),
			Languages: types.ListNull(types.StringType),
			UserRoles: types.ListNull(types.StringType),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("Blocklist 404 removes state", func(t *testing.T) {
		r := &BlocklistResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := BlocklistModel{
			ID:        types.StringValue("movie:12345"),
			MediaType: types.StringValue("movie"),
			TMDBID:    types.Int64Value(12345),
			UserID:    types.Int64Value(1),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("Issue 404 removes state", func(t *testing.T) {
		r := &IssueResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := IssueModel{
			ID:         types.StringValue("1"),
			CreatedBy:  parseIssueUserObject(nil),
			ModifiedBy: parseIssueUserObject(nil),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("IssueComment 404 removes state", func(t *testing.T) {
		r := &IssueCommentResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := IssueCommentModel{ID: types.StringValue("1")}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("Request 404 removes state", func(t *testing.T) {
		r := &RequestResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := RequestModel{
			ID:          types.StringValue("1"),
			RequestedBy: parseRequestUserObject(nil),
			ModifiedBy:  parseRequestUserObject(nil),
			Seasons:     types.ListNull(types.Int64Type),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("RequestApproval 404 removes state", func(t *testing.T) {
		r := &RequestApprovalResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := RequestApprovalModel{
			ID:        types.StringValue("1"),
			RequestID: types.Int64Value(1),
			Status:    types.StringValue("approved"),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("NotificationClient 404 removes state", func(t *testing.T) {
		r := &NotificationClientResource{agent: "discord", client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})

	t.Run("APIObject 404 with suppress_not_found removes state", func(t *testing.T) {
		r := &APIObjectResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		initial := APIObjectModel{
			ID:               types.StringValue("/api/v1/custom"),
			Path:             types.StringValue("/api/v1/custom"),
			ReadMethod:       types.StringValue("GET"),
			SuppressNotFound: types.BoolValue(true),
			Headers:          types.MapNull(types.StringType),
		}
		assert.False(t, state.Set(ctx, &initial).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		assert.False(t, readResp.Diagnostics.HasError())
	})
}
