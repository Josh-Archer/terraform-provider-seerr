package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/require"
)

func assertInvitationsUnsupported(t *testing.T, diagnostics diag.Diagnostics) {
	t.Helper()
	require.Len(t, diagnostics, 1)
	require.Equal(t, diag.SeverityError, diagnostics[0].Severity())
	require.Equal(t, "Invitations Unsupported", diagnostics[0].Summary())
	require.Equal(t, "Seerr does not expose an invitation API; use seerr_user to create local users", diagnostics[0].Detail())
}

func TestUserInvitationResourceUnsupported(t *testing.T) {
	ctx := context.Background()
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		http.NotFound(w, req)
	}))
	defer srv.Close()
	r, ok := NewUserInvitationResource().(*UserInvitationResource)
	require.True(t, ok)
	var configured resource.ConfigureResponse
	r.Configure(ctx, resource.ConfigureRequest{ProviderData: testAPIClient(t, srv.URL)}, &configured)
	require.False(t, configured.Diagnostics.HasError())
	var schema resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schema)
	require.Equal(t, invitationUnsupportedMessage, schema.Schema.DeprecationMessage)
	state := tfsdk.State{Schema: schema.Schema}
	require.False(t, state.Set(ctx, &UserInvitationModel{ID: types.StringValue("1"), Email: types.StringValue("user@example.com")}).HasError())
	plan := tfsdk.Plan{Schema: schema.Schema, Raw: state.Raw}
	t.Run("create", func(t *testing.T) {
		resp := resource.CreateResponse{}
		r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
		assertInvitationsUnsupported(t, resp.Diagnostics)
	})
	t.Run("read drops stale state with warning", func(t *testing.T) {
		resp := resource.ReadResponse{State: state}
		r.Read(ctx, resource.ReadRequest{State: state}, &resp)
		require.False(t, resp.Diagnostics.HasError())
		require.Len(t, resp.Diagnostics.Warnings(), 1)
		require.True(t, resp.State.Raw.IsNull())
	})
	t.Run("update", func(t *testing.T) {
		resp := resource.UpdateResponse{State: state}
		r.Update(ctx, resource.UpdateRequest{State: state, Plan: plan}, &resp)
		assertInvitationsUnsupported(t, resp.Diagnostics)
		require.True(t, state.Raw.Equal(resp.State.Raw))
	})
	t.Run("delete succeeds with warning", func(t *testing.T) {
		resp := resource.DeleteResponse{State: state}
		r.Delete(ctx, resource.DeleteRequest{State: state}, &resp)
		require.False(t, resp.Diagnostics.HasError())
		require.Len(t, resp.Diagnostics.Warnings(), 1)
	})
	t.Run("import", func(t *testing.T) {
		resp := resource.ImportStateResponse{}
		r.ImportState(ctx, resource.ImportStateRequest{ID: "1"}, &resp)
		assertInvitationsUnsupported(t, resp.Diagnostics)
	})
	require.Zero(t, requests.Load(), "unsupported operations must not make HTTP requests")
}
