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

type NotificationClientDataSourceModel struct {
	Enabled     types.Bool  `tfsdk:"enabled"`
	EmbedPoster types.Bool  `tfsdk:"embed_poster"`
	TypesMask   types.Int64 `tfsdk:"types"`

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
	OnIssueCreated        types.Bool                        `tfsdk:"on_issue_created"`
	OnIssueComment        types.Bool                        `tfsdk:"on_issue_comment"`
	OnIssueResolved       types.Bool                        `tfsdk:"on_issue_resolved"`
	OnIssueReopened       types.Bool                        `tfsdk:"on_issue_reopened"`
	OnMediaAutoRequested  types.Bool                        `tfsdk:"on_media_auto_requested"`
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

	data := NotificationClientDataSourceModel{}
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
	data.OnIssueCreated = resourceData.OnIssueCreated
	data.OnIssueComment = resourceData.OnIssueComment
	data.OnIssueResolved = resourceData.OnIssueResolved
	data.OnIssueReopened = resourceData.OnIssueReopened
	data.OnMediaAutoRequested = resourceData.OnMediaAutoRequested

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
