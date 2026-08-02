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

func TestPushbulletSettingsDataSourceMetadataAndSchema(t *testing.T) {
	t.Parallel()

	ds := NewPushbulletSettingsDataSource()
	require.NotNil(t, ds)

	metaResp := &datasource.MetadataResponse{}
	ds.Metadata(context.Background(), datasource.MetadataRequest{ProviderTypeName: "seerr"}, metaResp)
	assert.Equal(t, "seerr_pushbullet_settings", metaResp.TypeName)

	schemaResp := &datasource.SchemaResponse{}
	ds.Schema(context.Background(), datasource.SchemaRequest{}, schemaResp)
	require.False(t, schemaResp.Diagnostics.HasError())

	schema := schemaResp.Schema
	assert.Contains(t, schema.Attributes, "id")
	assert.Contains(t, schema.Attributes, "enabled")
	assert.Contains(t, schema.Attributes, "embed_poster")
	assert.Contains(t, schema.Attributes, "notification_types")
	assert.Contains(t, schema.Attributes, "pushbullet")
}

func TestPushbulletSettingsDataSourceRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/settings/notifications/pushbullet" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"enabled": true,
			"embedPoster": false,
			"types": 2,
			"options": {
				"accessToken": "o.mockAccessToken123",
				"channelTag": "seerr-channel"
			}
		}`))
	}))
	defer srv.Close()

	baseURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := NewClient(baseURL, "test-api-key", "test-agent", false, defaultRequestTimeout)

	ds, ok := newNotificationClientDataSourceWithTypeName("pushbullet", "pushbullet_settings").(*NotificationClientDataSource)
	require.True(t, ok)
	ds.client = client

	res, err := ds.client.Request(context.Background(), "GET", notificationPath("pushbullet"), "", nil)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var model NotificationAgentModel
	model.Agent = types.StringValue("pushbullet")
	err = parsePayload(context.Background(), &model, res.Body)
	require.NoError(t, err)

	assert.True(t, model.Enabled.ValueBool())
	assert.False(t, model.EmbedPoster.ValueBool())
	require.NotNil(t, model.Pushbullet)
	assert.Equal(t, "o.mockAccessToken123", model.Pushbullet.AccessToken.ValueString())
	assert.Equal(t, "seerr-channel", model.Pushbullet.ChannelTag.ValueString())
}
