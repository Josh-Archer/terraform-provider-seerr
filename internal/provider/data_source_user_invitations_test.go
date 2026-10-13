package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/require"
)

func TestUserInvitationsDataSourceUnsupported(t *testing.T) {
	ctx := context.Background()
	var requests atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		requests.Add(1)
		http.NotFound(w, req)
	}))
	defer srv.Close()
	d := NewUserInvitationsDataSource().(*UserInvitationsDataSource)
	var configured datasource.ConfigureResponse
	d.Configure(ctx, datasource.ConfigureRequest{ProviderData: testAPIClient(t, srv.URL)}, &configured)
	require.False(t, configured.Diagnostics.HasError())
	var schema datasource.SchemaResponse
	d.Schema(ctx, datasource.SchemaRequest{}, &schema)
	require.Equal(t, invitationUnsupportedMessage, schema.Schema.DeprecationMessage)
	var resp datasource.ReadResponse
	d.Read(ctx, datasource.ReadRequest{}, &resp)
	assertInvitationsUnsupported(t, resp.Diagnostics)
	require.Zero(t, requests.Load(), "unsupported operations must not make HTTP requests")
}
