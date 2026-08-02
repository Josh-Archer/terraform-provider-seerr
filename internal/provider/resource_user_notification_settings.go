package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserNotificationSettingsResource{}
var _ resource.ResourceWithImportState = &UserNotificationSettingsResource{}

type UserNotificationSettingsResource struct {
	client *APIClient
}

type UserNotificationSettingsResourceModel struct {
	ID                       types.String                `tfsdk:"id"`
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

func NewUserNotificationSettingsResource() resource.Resource {
	return &UserNotificationSettingsResource{}
}

func (r *UserNotificationSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_notification_settings"
}

func (r *UserNotificationSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage per-user notification settings in Seerr via `/api/v1/user/{userId}/settings/notifications`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of this resource (matches user_id).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the user whose notification settings to manage.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"email_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether email notifications are enabled for this user.",
				Optional:            true,
				Computed:            true,
			},
			"pgp_key": schema.StringAttribute{
				MarkdownDescription: "PGP public key for encrypted email notifications.",
				Optional:            true,
				Computed:            true,
			},
			"discord_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Discord notifications are enabled for this user.",
				Optional:            true,
				Computed:            true,
			},
			"discord_id": schema.StringAttribute{
				MarkdownDescription: "Discord user ID for direct notification mentions.",
				Optional:            true,
				Computed:            true,
			},
			"pushbullet_access_token": schema.StringAttribute{
				MarkdownDescription: "Per-user Pushbullet access token override.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"pushover_application_token": schema.StringAttribute{
				MarkdownDescription: "Per-user Pushover application token override.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"pushover_user_key": schema.StringAttribute{
				MarkdownDescription: "Per-user Pushover user key override.",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			"pushover_sound": schema.StringAttribute{
				MarkdownDescription: "Per-user Pushover notification sound override.",
				Optional:            true,
				Computed:            true,
			},
			"telegram_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Telegram notifications are enabled for this user.",
				Optional:            true,
				Computed:            true,
			},
			"telegram_bot_username": schema.StringAttribute{
				MarkdownDescription: "Telegram bot username for this user.",
				Optional:            true,
				Computed:            true,
			},
			"telegram_chat_id": schema.StringAttribute{
				MarkdownDescription: "Telegram chat ID for this user.",
				Optional:            true,
				Computed:            true,
			},
			"telegram_message_thread_id": schema.StringAttribute{
				MarkdownDescription: "Telegram message thread ID for topic-based group chats.",
				Optional:            true,
				Computed:            true,
			},
			"telegram_send_silently": schema.BoolAttribute{
				MarkdownDescription: "Whether Telegram notifications should be sent silently.",
				Optional:            true,
				Computed:            true,
			},
			"webpush_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Web Push notifications are enabled for this user.",
				Optional:            true,
				Computed:            true,
			},
			"notification_types": schema.SingleNestedAttribute{
				MarkdownDescription: "Per-user notification type bitmasks per agent.",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"discord":    schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Discord.", Optional: true, Computed: true},
					"email":      schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Email.", Optional: true, Computed: true},
					"pushbullet": schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Pushbullet.", Optional: true, Computed: true},
					"pushover":   schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Pushover.", Optional: true, Computed: true},
					"slack":      schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Slack.", Optional: true, Computed: true},
					"telegram":   schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Telegram.", Optional: true, Computed: true},
					"webhook":    schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Webhook.", Optional: true, Computed: true},
					"webpush":    schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Webpush.", Optional: true, Computed: true},
					"gotify":     schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Gotify.", Optional: true, Computed: true},
					"ntfy":       schema.Int64Attribute{MarkdownDescription: "Notification types bitmask for Ntfy.", Optional: true, Computed: true},
				},
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body from the latest operation.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code from the latest operation.",
				Computed:            true,
			},
		},
	}
}

func (r *UserNotificationSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func userNotificationSettingsPath(userID int64) string {
	return fmt.Sprintf("/api/v1/user/%d/settings/notifications", userID)
}

func (r *UserNotificationSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserNotificationSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildUserNotificationSettingsPayload(ctx, r.client, data.UserID.ValueInt64(), &data)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	apiPath := userNotificationSettingsPath(data.UserID.ValueInt64())
	res, err := r.client.Request(ctx, "POST", apiPath, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	if err := r.readNotificationSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserNotificationSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserNotificationSettingsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readNotificationSettings(ctx, &data); err != nil {
		if err.Error() == "not found" {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserNotificationSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserNotificationSettingsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload, err := buildUserNotificationSettingsPayload(ctx, r.client, data.UserID.ValueInt64(), &data)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	apiPath := userNotificationSettingsPath(data.UserID.ValueInt64())
	res, err := r.client.Request(ctx, "POST", apiPath, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	if err := r.readNotificationSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserNotificationSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No DELETE route for user notification settings; removing from state only.
}

func (r *UserNotificationSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := requireInt64ID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), id)...)
}

func (r *UserNotificationSettingsResource) readNotificationSettings(ctx context.Context, data *UserNotificationSettingsResourceModel) error {
	userID := data.UserID.ValueInt64()
	apiPath := userNotificationSettingsPath(userID)
	res, err := r.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		return err
	}
	if res.StatusCode == 404 {
		return fmt.Errorf("not found")
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
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

	data.ID = types.StringValue(fmt.Sprintf("%d", userID))
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	return nil
}

func populateUserNotificationSettingsFromMap(
	decoded map[string]any,
	emailEnabled *types.Bool,
	pgpKey *types.String,
	discordEnabled *types.Bool,
	discordID *types.String,
	pushbulletAccessToken *types.String,
	pushoverApplicationToken *types.String,
	pushoverUserKey *types.String,
	pushoverSound *types.String,
	telegramEnabled *types.Bool,
	telegramBotUsername *types.String,
	telegramChatID *types.String,
	telegramMessageThreadID *types.String,
	telegramSendSilently *types.Bool,
	webpushEnabled *types.Bool,
	notificationTypes **UserNotificationTypesModel,
) {
	if v, ok := boolValueFromAny(decoded["emailEnabled"]); ok {
		*emailEnabled = types.BoolValue(v)
	}
	if v, ok := stringValueFromAny(decoded["pgpKey"]); ok {
		*pgpKey = types.StringValue(v)
	}
	if v, ok := boolValueFromAny(decoded["discordEnabled"]); ok {
		*discordEnabled = types.BoolValue(v)
	}
	if v, ok := stringValueFromAny(decoded["discordId"]); ok {
		*discordID = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["pushbulletAccessToken"]); ok {
		*pushbulletAccessToken = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["pushoverApplicationToken"]); ok {
		*pushoverApplicationToken = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["pushoverUserKey"]); ok {
		*pushoverUserKey = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["pushoverSound"]); ok {
		*pushoverSound = types.StringValue(v)
	}
	if v, ok := boolValueFromAny(decoded["telegramEnabled"]); ok {
		*telegramEnabled = types.BoolValue(v)
	}
	if v, ok := stringValueFromAny(decoded["telegramBotUsername"]); ok {
		*telegramBotUsername = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["telegramChatId"]); ok {
		*telegramChatID = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["telegramMessageThreadId"]); ok {
		*telegramMessageThreadID = types.StringValue(v)
	}
	if v, ok := boolValueFromAny(decoded["telegramSendSilently"]); ok {
		*telegramSendSilently = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["webpushEnabled"]); ok {
		*webpushEnabled = types.BoolValue(v)
	} else if v, ok := boolValueFromAny(decoded["webPushEnabled"]); ok {
		*webpushEnabled = types.BoolValue(v)
	}

	if typesMap, ok := decoded["notificationTypes"].(map[string]any); ok {
		nt := &UserNotificationTypesModel{}
		if v, ok := int64ValueFromAny(typesMap["discord"]); ok {
			nt.Discord = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(typesMap["email"]); ok {
			nt.Email = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(typesMap["pushbullet"]); ok {
			nt.Pushbullet = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(typesMap["pushover"]); ok {
			nt.Pushover = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(typesMap["slack"]); ok {
			nt.Slack = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(typesMap["telegram"]); ok {
			nt.Telegram = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(typesMap["webhook"]); ok {
			nt.Webhook = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(typesMap["webpush"]); ok {
			nt.Webpush = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(typesMap["gotify"]); ok {
			nt.Gotify = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(typesMap["ntfy"]); ok {
			nt.Ntfy = types.Int64Value(v)
		}
		*notificationTypes = nt
	}
}

func buildUserNotificationSettingsPayload(ctx context.Context, client *APIClient, userID int64, data *UserNotificationSettingsResourceModel) (map[string]any, error) {
	if client == nil {
		return nil, fmt.Errorf("API client is nil")
	}
	apiPath := userNotificationSettingsPath(userID)
	res, err := client.Request(ctx, "GET", apiPath, "", nil)
	var current map[string]any
	if err == nil && StatusIsOK(res.StatusCode) {
		_ = json.Unmarshal(res.Body, &current)
	}
	if current == nil {
		current = map[string]any{}
	}

	payload := copyMap(current)
	setOptionalBool(payload, "emailEnabled", data.EmailEnabled)
	setOptionalString(payload, "pgpKey", data.PGPKey)
	setOptionalBool(payload, "discordEnabled", data.DiscordEnabled)
	setOptionalString(payload, "discordId", data.DiscordID)
	setOptionalString(payload, "pushbulletAccessToken", data.PushbulletAccessToken)
	setOptionalString(payload, "pushoverApplicationToken", data.PushoverApplicationToken)
	setOptionalString(payload, "pushoverUserKey", data.PushoverUserKey)
	setOptionalString(payload, "pushoverSound", data.PushoverSound)
	setOptionalBool(payload, "telegramEnabled", data.TelegramEnabled)
	setOptionalString(payload, "telegramBotUsername", data.TelegramBotUsername)
	setOptionalString(payload, "telegramChatId", data.TelegramChatID)
	setOptionalString(payload, "telegramMessageThreadId", data.TelegramMessageThreadID)
	setOptionalBool(payload, "telegramSendSilently", data.TelegramSendSilently)
	setOptionalBool(payload, "webpushEnabled", data.WebpushEnabled)

	if data.NotificationTypes != nil {
		existingTypes := map[string]any{}
		if rawExisting, ok := payload["notificationTypes"].(map[string]any); ok {
			existingTypes = copyMap(rawExisting)
		}

		setOptionalInt64(existingTypes, "discord", data.NotificationTypes.Discord)
		setOptionalInt64(existingTypes, "email", data.NotificationTypes.Email)
		setOptionalInt64(existingTypes, "pushbullet", data.NotificationTypes.Pushbullet)
		setOptionalInt64(existingTypes, "pushover", data.NotificationTypes.Pushover)
		setOptionalInt64(existingTypes, "slack", data.NotificationTypes.Slack)
		setOptionalInt64(existingTypes, "telegram", data.NotificationTypes.Telegram)
		setOptionalInt64(existingTypes, "webhook", data.NotificationTypes.Webhook)
		setOptionalInt64(existingTypes, "webpush", data.NotificationTypes.Webpush)
		setOptionalInt64(existingTypes, "gotify", data.NotificationTypes.Gotify)
		setOptionalInt64(existingTypes, "ntfy", data.NotificationTypes.Ntfy)

		payload["notificationTypes"] = existingTypes
	}

	return payload, nil
}
