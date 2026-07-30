package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserNotificationSettingsDataSource{}

type UserNotificationSettingsDataSource struct {
	client *APIClient
}

type UserNotificationSettingsDataSourceModel struct {
	UserID                   types.Int64                 `tfsdk:"user_id"`
	EmailEnabled             types.Bool                  `tfsdk:"email_enabled"`
	PGPKey                   types.String                `tfsdk:"pgp_key"`
	DiscordEnabled           types.Bool                  `tfsdk:"discord_enabled"`
	DiscordID                types.String                `tfsdk:"discord_id"`
	PushbulletAccessToken    types.String                `tfsdk:"pushbullet_access_token"`
	PushoverApplicationToken types.String                `tfsdk:"pushover_application_token"`
	PushoverUserKey          types.String                `tfsdk:"pushover_user_key"`
	PushoverSound            types.String                `tfsdk:"pushover_sound"`
	TelegramEnabled          types.Bool                  `tfsdk:"telegram_enabled"`
	TelegramBotUsername      types.String                `tfsdk:"telegram_bot_username"`
	TelegramChatID           types.String                `tfsdk:"telegram_chat_id"`
	TelegramMessageThreadID  types.String                `tfsdk:"telegram_message_thread_id"`
	TelegramSendSilently     types.Bool                  `tfsdk:"telegram_send_silently"`
	WebpushEnabled           types.Bool                  `tfsdk:"webpush_enabled"`
	NotificationTypes        *UserNotificationTypesModel `tfsdk:"notification_types"`
	ResponseJSON             types.String                `tfsdk:"response_json"`
	StatusCode               types.Int64                 `tfsdk:"status_code"`
}

func NewUserNotificationSettingsDataSource() datasource.DataSource {
	return &UserNotificationSettingsDataSource{}
}

func (d *UserNotificationSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_notification_settings"
}

func (d *UserNotificationSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		MarkdownDescription: "Read per-user notification settings from Seerr via `/api/v1/user/{userId}/settings/notifications`.",
		Attributes: map[string]dschema.Attribute{
			"user_id": dschema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the user to look up.",
				Required:            true,
			},
			"email_enabled": dschema.BoolAttribute{
				MarkdownDescription: "Whether email notifications are enabled for this user.",
				Computed:            true,
			},
			"pgp_key": dschema.StringAttribute{
				MarkdownDescription: "PGP public key for encrypted email notifications.",
				Computed:            true,
			},
			"discord_enabled": dschema.BoolAttribute{
				MarkdownDescription: "Whether Discord notifications are enabled for this user.",
				Computed:            true,
			},
			"discord_id": dschema.StringAttribute{
				MarkdownDescription: "Discord user ID for direct notification mentions.",
				Computed:            true,
			},
			"pushbullet_access_token": dschema.StringAttribute{
				MarkdownDescription: "Per-user Pushbullet access token override.",
				Computed:            true,
				Sensitive:           true,
			},
			"pushover_application_token": dschema.StringAttribute{
				MarkdownDescription: "Per-user Pushover application token override.",
				Computed:            true,
				Sensitive:           true,
			},
			"pushover_user_key": dschema.StringAttribute{
				MarkdownDescription: "Per-user Pushover user key override.",
				Computed:            true,
				Sensitive:           true,
			},
			"pushover_sound": dschema.StringAttribute{
				MarkdownDescription: "Per-user Pushover notification sound override.",
				Computed:            true,
			},
			"telegram_enabled": dschema.BoolAttribute{
				MarkdownDescription: "Whether Telegram notifications are enabled for this user.",
				Computed:            true,
			},
			"telegram_bot_username": dschema.StringAttribute{
				MarkdownDescription: "Telegram bot username for this user.",
				Computed:            true,
			},
			"telegram_chat_id": dschema.StringAttribute{
				MarkdownDescription: "Telegram chat ID for this user.",
				Computed:            true,
			},
			"telegram_message_thread_id": dschema.StringAttribute{
				MarkdownDescription: "Telegram message thread ID for topic-based group chats.",
				Computed:            true,
			},
			"telegram_send_silently": dschema.BoolAttribute{
				MarkdownDescription: "Whether Telegram notifications should be sent silently.",
				Computed:            true,
			},
			"webpush_enabled": dschema.BoolAttribute{
				MarkdownDescription: "Whether Web Push notifications are enabled for this user.",
				Computed:            true,
			},
			"notification_types": dschema.SingleNestedAttribute{
				MarkdownDescription: "Per-user notification type bitmasks per agent.",
				Computed:            true,
				Attributes: map[string]dschema.Attribute{
					"discord":    dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Discord.", Computed: true},
					"email":      dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Email.", Computed: true},
					"pushbullet": dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Pushbullet.", Computed: true},
					"pushover":   dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Pushover.", Computed: true},
					"slack":      dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Slack.", Computed: true},
					"telegram":   dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Telegram.", Computed: true},
					"webhook":    dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Webhook.", Computed: true},
					"webpush":    dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Webpush.", Computed: true},
					"gotify":     dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Gotify.", Computed: true},
					"ntfy":       dschema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Ntfy.", Computed: true},
				},
			},
			"response_json": dschema.StringAttribute{
				MarkdownDescription: "Raw JSON response body.",
				Computed:            true,
			},
			"status_code": dschema.Int64Attribute{
				MarkdownDescription: "HTTP status code.",
				Computed:            true,
			},
		},
	}
}

func (d *UserNotificationSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserNotificationSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserNotificationSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := userNotificationSettingsPath(data.UserID.ValueInt64())
	res, err := d.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err.Error()))
		return
	}

	populateUserNotificationSettingsFromMap(
		decoded,
		&data.EmailEnabled,
		&data.PGPKey,
		&data.DiscordEnabled,
		&data.DiscordID,
		&data.PushbulletAccessToken,
		&data.PushoverApplicationToken,
		&data.PushoverUserKey,
		&data.PushoverSound,
		&data.TelegramEnabled,
		&data.TelegramBotUsername,
		&data.TelegramChatID,
		&data.TelegramMessageThreadID,
		&data.TelegramSendSilently,
		&data.WebpushEnabled,
		&data.NotificationTypes,
	)

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
