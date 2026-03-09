package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationClientDataSource struct {
	client *APIClient
	agent  string
}

var _ datasource.DataSource = &NotificationClientDataSource{}

func newNotificationClientDataSource(agent string) datasource.DataSource {
	return &NotificationClientDataSource{agent: agent}
}

func NewNotificationDiscordDataSource() datasource.DataSource {
	return newNotificationClientDataSource("discord")
}
func NewNotificationSlackDataSource() datasource.DataSource {
	return newNotificationClientDataSource("slack")
}
func NewNotificationEmailDataSource() datasource.DataSource {
	return newNotificationClientDataSource("email")
}
func NewNotificationLunaSeaDataSource() datasource.DataSource {
	return newNotificationClientDataSource("lunasea")
}
func NewNotificationTelegramDataSource() datasource.DataSource {
	return newNotificationClientDataSource("telegram")
}
func NewNotificationPushbulletDataSource() datasource.DataSource {
	return newNotificationClientDataSource("pushbullet")
}
func NewNotificationPushoverDataSource() datasource.DataSource {
	return newNotificationClientDataSource("pushover")
}
func NewNotificationNtfyDataSource() datasource.DataSource {
	return newNotificationClientDataSource("ntfy")
}
func NewNotificationWebhookDataSource() datasource.DataSource {
	return newNotificationClientDataSource("webhook")
}
func NewNotificationGotifyDataSource() datasource.DataSource {
	return newNotificationClientDataSource("gotify")
}
func NewNotificationWebpushDataSource() datasource.DataSource {
	return newNotificationClientDataSource("webpush")
}

func (d *NotificationClientDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_" + d.agent
}

func (d *NotificationClientDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"enabled":      schema.BoolAttribute{Computed: true},
		"embed_poster": schema.BoolAttribute{Computed: true},
		"types":        schema.Int64Attribute{Computed: true},
	}
	for name, attr := range notificationAgentDataSourceEventAttributes() {
		attributes[name] = attr
	}

	optionAttr, ok := notificationAgentDataSourceOptionAttribute(d.agent)
	if !ok {
		resp.Diagnostics.AddError("Unsupported notification agent", fmt.Sprintf("agent %q is not supported", d.agent))
		return
	}
	attributes[d.agent] = optionAttr

	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf("Read Seerr %s notification settings via /api/v1/settings/notifications/%s.", d.agent, d.agent),
		Attributes:          attributes,
	}
}

func (d *NotificationClientDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *NotificationClientDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	res, err := d.client.Request(ctx, "GET", notificationPath(d.agent), "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	resourceData := NotificationAgentModel{Agent: types.StringValue(d.agent)}
	if err := parsePayload(&resourceData, res.Body); err != nil {
		resp.Diagnostics.AddError("Parse Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(setNotificationClientState(ctx, &resp.State, &resourceData)...)
}
