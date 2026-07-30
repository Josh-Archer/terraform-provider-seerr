package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailSettingsDataSourceMetadataAndSchema(t *testing.T) {
	t.Parallel()

	ds := NewEmailSettingsDataSource()
	require.NotNil(t, ds)

	metaResp := &datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, metaResp)
	assert.Equal(t, "seerr_email_settings", metaResp.TypeName)

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	schema := schemaResp.Schema
	assert.Contains(t, schema.Attributes, "id")
	assert.Contains(t, schema.Attributes, "enabled")
	assert.Contains(t, schema.Attributes, "embed_poster")
	assert.Contains(t, schema.Attributes, "notification_types")
	assert.Contains(t, schema.Attributes, "email")
}

func TestEmailSettingsDataSourceRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/settings/notifications/email" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"enabled": true,
			"embedPoster": true,
			"types": 6,
			"options": {
				"emailFrom": "no-reply@example.com",
				"smtpHost": "smtp.example.com",
				"smtpPort": 587,
				"secure": true,
				"ignoreTls": false,
				"requireTls": true,
				"authUser": "smtpuser",
				"authPass": "smtppass",
				"allowSelfSigned": false,
				"senderName": "Seerr Notifications",
				"pgpPrivateKey": "pgp-key-data",
				"pgpPassword": "pgp-pass"
			}
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout)

	ds := newNotificationClientDataSourceWithTypeName("email", "email_settings").(*NotificationClientDataSource)
	ds.client = client

	res, err := ds.client.Request(context.Background(), "GET", notificationPath("email"), "", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var model NotificationAgentModel
	model.Agent = types.StringValue("email")
	err = parsePayload(context.Background(), &model, res.Body)
	require.NoError(t, err)

	assert.True(t, model.Enabled.ValueBool())
	assert.True(t, model.EmbedPoster.ValueBool())
	require.NotNil(t, model.Email)
	assert.Equal(t, "no-reply@example.com", model.Email.EmailFrom.ValueString())
	assert.Equal(t, "smtp.example.com", model.Email.SmtpHost.ValueString())
	assert.Equal(t, int64(587), model.Email.SmtpPort.ValueInt64())
	assert.True(t, model.Email.Secure.ValueBool())
	assert.False(t, model.Email.IgnoreTls.ValueBool())
	assert.True(t, model.Email.RequireTls.ValueBool())
	assert.Equal(t, "smtpuser", model.Email.AuthUser.ValueString())
	assert.Equal(t, "smtppass", model.Email.AuthPass.ValueString())
	assert.False(t, model.Email.AllowSelfSigned.ValueBool())
	assert.Equal(t, "Seerr Notifications", model.Email.SenderName.ValueString())
	assert.Equal(t, "pgp-key-data", model.Email.PgpPrivateKey.ValueString())
	assert.Equal(t, "pgp-pass", model.Email.PgpPassword.ValueString())
}
