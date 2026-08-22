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
	"github.com/stretchr/testify/require"
)

func TestUserWatchlistPathUsesMainSettingsEndpoint(t *testing.T) {
	if got, want := userWatchlistPath(1), "/api/v1/user/1/settings/main"; got != want {
		t.Fatalf("userWatchlistPath(1) = %q, want %q", got, want)
	}
}

func TestResource404StateRemoval(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Not Found"}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)
	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout, 0, 0)
	ctx := context.Background()

	t.Run("UserResource", func(t *testing.T) {
		r := &UserResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		require.False(t, state.Set(ctx, &UserModel{
			ID:    types.StringValue("99"),
			Email: types.StringValue("user@example.com"),
		}).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		require.False(t, readResp.Diagnostics.HasError(), "404 Read should not return errors")
		assert.True(t, readResp.State.Raw.IsNull(), "State should be removed (Null) after 404")
	})

	t.Run("UserPermissionsResource", func(t *testing.T) {
		r := &UserPermissionsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		require.False(t, state.Set(ctx, &UserPermissionsModel{
			ID:     types.StringValue("99"),
			UserID: types.Int64Value(99),
		}).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		require.False(t, readResp.Diagnostics.HasError(), "404 Read should not return errors")
		assert.True(t, readResp.State.Raw.IsNull(), "State should be removed (Null) after 404")
	})

	t.Run("UserQuotaResource", func(t *testing.T) {
		r := &UserQuotaResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		require.False(t, state.Set(ctx, &UserQuotaModel{
			ID:     types.StringValue("99"),
			UserID: types.Int64Value(99),
		}).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		require.False(t, readResp.Diagnostics.HasError(), "404 Read should not return errors")
		assert.True(t, readResp.State.Raw.IsNull(), "State should be removed (Null) after 404")
	})

	t.Run("UserSettingsPermissionsResource", func(t *testing.T) {
		r := &UserSettingsPermissionsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		require.False(t, state.Set(ctx, &UserSettingsPermissionsModel{
			ID:     types.StringValue("99"),
			UserID: types.Int64Value(99),
		}).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		require.False(t, readResp.Diagnostics.HasError(), "404 Read should not return errors")
		assert.True(t, readResp.State.Raw.IsNull(), "State should be removed (Null) after 404")
	})

	t.Run("UserWatchlistSettingsResource", func(t *testing.T) {
		r := &UserWatchlistSettingsResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		require.False(t, state.Set(ctx, &UserWatchlistSettingsModel{
			ID:     types.StringValue("99"),
			UserID: types.Int64Value(99),
		}).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		require.False(t, readResp.Diagnostics.HasError(), "404 Read should not return errors")
		assert.True(t, readResp.State.Raw.IsNull(), "State should be removed (Null) after 404")
	})

	t.Run("RequestApprovalResource", func(t *testing.T) {
		r := &RequestApprovalResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		require.False(t, state.Set(ctx, &RequestApprovalModel{
			ID:        types.StringValue("99"),
			RequestID: types.Int64Value(99),
		}).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		require.False(t, readResp.Diagnostics.HasError(), "404 Read should not return errors")
		assert.True(t, readResp.State.Raw.IsNull(), "State should be removed (Null) after 404")
	})

	t.Run("RequestRetryResource", func(t *testing.T) {
		r := &RequestRetryResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		require.False(t, state.Set(ctx, &RequestRetryModel{
			ID:        types.StringValue("99"),
			RequestID: types.Int64Value(99),
		}).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		require.False(t, readResp.Diagnostics.HasError(), "404 Read should not return errors")
		assert.True(t, readResp.State.Raw.IsNull(), "State should be removed (Null) after 404")
	})

	t.Run("BlocklistResource", func(t *testing.T) {
		r := &BlocklistResource{client: client}
		var schemaResp resource.SchemaResponse
		r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		state := tfsdk.State{Schema: schemaResp.Schema}
		require.False(t, state.Set(ctx, &BlocklistModel{
			ID:        types.StringValue("movie:99"),
			TMDBID:    types.Int64Value(99),
			MediaType: types.StringValue("movie"),
		}).HasError())

		readResp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &readResp)
		require.False(t, readResp.Diagnostics.HasError(), "404 Read should not return errors")
		assert.True(t, readResp.State.Raw.IsNull(), "State should be removed (Null) after 404")
	})
}
