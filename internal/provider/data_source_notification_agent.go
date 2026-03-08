package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NotificationAgentDataSource{}

type NotificationAgentDataSource struct {
	client *APIClient
}

type NotificationAgentDataSourceModel struct {
	Agent       types.String `tfsdk:"agent"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	EmbedPoster types.Bool   `tfsdk:"embed_poster"`
	TypesMask   types.Int64  `tfsdk:"types"`

	Discord               *NotificationAgentDiscordModel    `tfsdk:"discord"`
	Slack                 *NotificationAgentSlackModel      `tfsdk:"slack"`
	Email                 *NotificationAgentEmailModel      `tfsdk:"email"`
	LunaSea               *NotificationAgentLunaSeaModel    `tfsdk:"lunasea"`
	Telegram              *NotificationAgentTelegramModel   `tfsdk:"telegram"`
	Pushbullet            *NotificationAgentPushbulletModel `tfsdk:"pushbullet"`
	Pushover              *NotificationAgentPushoverModel   `tfsdk:"pushover"`
	Ntfy                  *NotificationAgentNtfyModel       `tfsdk:"ntfy"`
	Webhook               *NotificationAgentWebhookModel    `tfsdk:"webhook"`
	Gotify                *NotificationAgentGotifyModel     `tfsdk:"gotify"`
	Webpush               *NotificationAgentWebpushModel    `tfsdk:"webpush"`
	OnRequestPending      types.Bool                        `tfsdk:"on_request_pending"`
	OnRequestApproved     types.Bool                        `tfsdk:"on_request_approved"`
	OnRequestRejected     types.Bool                        `tfsdk:"on_request_rejected"`
	OnRequestFailed       types.Bool                        `tfsdk:"on_request_failed"`
	OnRequestAvailable    types.Bool                        `tfsdk:"on_request_available"`
	OnRequestDeclined     types.Bool                        `tfsdk:"on_request_declined"`
	OnRequestAutoApproved types.Bool                        `tfsdk:"on_request_auto_approved"`
	OnMediaAvailable      types.Bool                        `tfsdk:"on_media_available"`
	OnMediaFailed         types.Bool                        `tfsdk:"on_media_failed"`
	OnMediaSkipped        types.Bool                        `tfsdk:"on_media_skipped"`
	OnMediaIssued         types.Bool                        `tfsdk:"on_media_issued"`
	OnMediaFollowed       types.Bool                        `tfsdk:"on_media_followed"`
}

func NewNotificationAgentDataSource() datasource.DataSource { return &NotificationAgentDataSource{} }

func (d *NotificationAgentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_agent"
}

func (d *NotificationAgentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"agent": schema.StringAttribute{
			MarkdownDescription: "Notification agent name (e.g. `email`, `discord`, `slack`).",
			Required:            true,
		},
		"enabled": schema.BoolAttribute{
			Computed: true,
		},
		"embed_poster": schema.BoolAttribute{
			Computed: true,
		},
		"types": schema.Int64Attribute{
			Computed: true,
		},
	}
	for name, attr := range notificationAgentDataSourceAttributes() {
		attributes[name] = attr
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Seerr notification agent settings via /api/v1/settings/notifications/{agent}.",
		Attributes:          attributes,
	}
}

func (d *NotificationAgentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NotificationAgentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationAgentDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := notificationPath(data.Agent.ValueString())
	res, err := d.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	// Reuse the resource model parsing logic by adapting types
	var resourceData NotificationAgentModel
	resourceData.Agent = data.Agent
	if err := parsePayload(&resourceData, res.Body); err != nil {
		resp.Diagnostics.AddError("Parse Failed", err.Error())
		return
	}

	data.Enabled = resourceData.Enabled
	data.EmbedPoster = resourceData.EmbedPoster
	data.TypesMask = resourceData.TypesMask
	data.Discord = resourceData.Discord
	data.Slack = resourceData.Slack
	data.Email = resourceData.Email
	data.LunaSea = resourceData.LunaSea
	data.Telegram = resourceData.Telegram
	data.Pushbullet = resourceData.Pushbullet
	data.Pushover = resourceData.Pushover
	data.Ntfy = resourceData.Ntfy
	data.Webhook = resourceData.Webhook
	data.Gotify = resourceData.Gotify
	data.Webpush = resourceData.Webpush

	data.OnRequestPending = resourceData.OnRequestPending
	data.OnRequestApproved = resourceData.OnRequestApproved
	data.OnRequestRejected = resourceData.OnRequestRejected
	data.OnRequestFailed = resourceData.OnRequestFailed
	data.OnRequestAvailable = resourceData.OnRequestAvailable
	data.OnRequestDeclined = resourceData.OnRequestDeclined
	data.OnRequestAutoApproved = resourceData.OnRequestAutoApproved
	data.OnMediaAvailable = resourceData.OnMediaAvailable
	data.OnMediaFailed = resourceData.OnMediaFailed
	data.OnMediaSkipped = resourceData.OnMediaSkipped
	data.OnMediaIssued = resourceData.OnMediaIssued
	data.OnMediaFollowed = resourceData.OnMediaFollowed

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
