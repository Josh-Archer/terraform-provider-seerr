package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &NotificationAgentResource{}
var _ resource.ResourceWithImportState = &NotificationAgentResource{}

type NotificationAgentResource struct {
	client *APIClient
}

type NotificationAgentModel struct {
	ID          types.String `tfsdk:"id"`
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

type notificationAgentPayload struct {
	Enabled     bool                   `json:"enabled"`
	EmbedPoster bool                   `json:"embedPoster"`
	Types       int64                  `json:"types"`
	Options     map[string]interface{} `json:"options"`
}

func NewNotificationAgentResource() resource.Resource { return &NotificationAgentResource{} }

func (r *NotificationAgentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_agent"
}

func (r *NotificationAgentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
			},
		},
		"agent": schema.StringAttribute{
			Required: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"enabled": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"embed_poster": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(false),
		},
		"types": schema.Int64Attribute{
			Optional: true,
			Computed: true,
			Default:  int64default.StaticInt64(0),
		},
	}
	for name, attr := range notificationAgentResourceAttributes() {
		attributes[name] = attr
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr notification agent settings via /api/v1/settings/notifications/{agent}.",
		Attributes:          attributes,
	}
}

func (r *NotificationAgentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func notificationPath(agent string) string {
	return "/api/v1/settings/notifications/" + agent
}

func buildPayload(data *NotificationAgentModel) (string, error) {
	payload := notificationAgentPayload{
		Enabled:     data.Enabled.ValueBool(),
		EmbedPoster: data.EmbedPoster.ValueBool(),
		Types:       data.TypesMask.ValueInt64(),
		Options:     make(map[string]interface{}),
	}

	// Calculate types bitmask from individual booleans if they are set
	mask := data.TypesMask.ValueInt64()
	if !data.OnRequestPending.IsNull() && !data.OnRequestPending.IsUnknown() {
		if data.OnRequestPending.ValueBool() {
			mask |= 1
		} else {
			mask &= ^int64(1)
		}
	}
	if !data.OnRequestApproved.IsNull() && !data.OnRequestApproved.IsUnknown() {
		if data.OnRequestApproved.ValueBool() {
			mask |= 2
		} else {
			mask &= ^int64(2)
		}
	}
	if !data.OnRequestRejected.IsNull() && !data.OnRequestRejected.IsUnknown() {
		if data.OnRequestRejected.ValueBool() {
			mask |= 4
		} else {
			mask &= ^int64(4)
		}
	}
	if !data.OnRequestFailed.IsNull() && !data.OnRequestFailed.IsUnknown() {
		if data.OnRequestFailed.ValueBool() {
			mask |= 8
		} else {
			mask &= ^int64(8)
		}
	}
	if !data.OnRequestAvailable.IsNull() && !data.OnRequestAvailable.IsUnknown() {
		if data.OnRequestAvailable.ValueBool() {
			mask |= 16
		} else {
			mask &= ^int64(16)
		}
	}
	if !data.OnRequestDeclined.IsNull() && !data.OnRequestDeclined.IsUnknown() {
		if data.OnRequestDeclined.ValueBool() {
			mask |= 32
		} else {
			mask &= ^int64(32)
		}
	}
	if !data.OnRequestAutoApproved.IsNull() && !data.OnRequestAutoApproved.IsUnknown() {
		if data.OnRequestAutoApproved.ValueBool() {
			mask |= 64
		} else {
			mask &= ^int64(64)
		}
	}
	if !data.OnMediaAvailable.IsNull() && !data.OnMediaAvailable.IsUnknown() {
		if data.OnMediaAvailable.ValueBool() {
			mask |= 128
		} else {
			mask &= ^int64(128)
		}
	}
	if !data.OnMediaFailed.IsNull() && !data.OnMediaFailed.IsUnknown() {
		if data.OnMediaFailed.ValueBool() {
			mask |= 256
		} else {
			mask &= ^int64(256)
		}
	}
	if !data.OnMediaSkipped.IsNull() && !data.OnMediaSkipped.IsUnknown() {
		if data.OnMediaSkipped.ValueBool() {
			mask |= 512
		} else {
			mask &= ^int64(512)
		}
	}
	if !data.OnMediaIssued.IsNull() && !data.OnMediaIssued.IsUnknown() {
		if data.OnMediaIssued.ValueBool() {
			mask |= 1024
		} else {
			mask &= ^int64(1024)
		}
	}
	if !data.OnMediaFollowed.IsNull() && !data.OnMediaFollowed.IsUnknown() {
		if data.OnMediaFollowed.ValueBool() {
			mask |= 2048
		} else {
			mask &= ^int64(2048)
		}
	}
	payload.Types = mask

	switch data.Agent.ValueString() {
	case "discord":
		if data.Discord != nil {
			if !data.Discord.BotUsername.IsNull() {
				payload.Options["botUsername"] = data.Discord.BotUsername.ValueString()
			}
			if !data.Discord.BotAvatarUrl.IsNull() {
				payload.Options["botAvatarUrl"] = data.Discord.BotAvatarUrl.ValueString()
			}
			if !data.Discord.WebhookUrl.IsNull() {
				payload.Options["webhookUrl"] = data.Discord.WebhookUrl.ValueString()
			}
			if !data.Discord.EnableMentions.IsNull() {
				payload.Options["enableMentions"] = data.Discord.EnableMentions.ValueBool()
			}
		}
	case "slack":
		if data.Slack != nil {
			if !data.Slack.WebhookUrl.IsNull() {
				payload.Options["webhookUrl"] = data.Slack.WebhookUrl.ValueString()
			}
		}
	case "email":
		if data.Email != nil {
			if !data.Email.EmailFrom.IsNull() {
				payload.Options["emailFrom"] = data.Email.EmailFrom.ValueString()
			}
			if !data.Email.SmtpHost.IsNull() {
				payload.Options["smtpHost"] = data.Email.SmtpHost.ValueString()
			}
			if !data.Email.SmtpPort.IsNull() {
				payload.Options["smtpPort"] = data.Email.SmtpPort.ValueInt64()
			}
			if !data.Email.Secure.IsNull() {
				payload.Options["secure"] = data.Email.Secure.ValueBool()
			}
			if !data.Email.IgnoreTls.IsNull() {
				payload.Options["ignoreTls"] = data.Email.IgnoreTls.ValueBool()
			}
			if !data.Email.RequireTls.IsNull() {
				payload.Options["requireTls"] = data.Email.RequireTls.ValueBool()
			}
			if !data.Email.AuthUser.IsNull() {
				payload.Options["authUser"] = data.Email.AuthUser.ValueString()
			}
			if !data.Email.AuthPass.IsNull() {
				payload.Options["authPass"] = data.Email.AuthPass.ValueString()
			}
			if !data.Email.AllowSelfSigned.IsNull() {
				payload.Options["allowSelfSigned"] = data.Email.AllowSelfSigned.ValueBool()
			}
			if !data.Email.SenderName.IsNull() {
				payload.Options["senderName"] = data.Email.SenderName.ValueString()
			}
			if !data.Email.PgpPrivateKey.IsNull() {
				payload.Options["pgpPrivateKey"] = data.Email.PgpPrivateKey.ValueString()
			}
			if !data.Email.PgpPassword.IsNull() {
				payload.Options["pgpPassword"] = data.Email.PgpPassword.ValueString()
			}
		}
	case "lunasea":
		if data.LunaSea != nil {
			if !data.LunaSea.WebhookUrl.IsNull() {
				payload.Options["webhookUrl"] = data.LunaSea.WebhookUrl.ValueString()
			}
			if !data.LunaSea.ProfileName.IsNull() {
				payload.Options["profileName"] = data.LunaSea.ProfileName.ValueString()
			}
		}
	case "telegram":
		if data.Telegram != nil {
			if !data.Telegram.BotUsername.IsNull() {
				payload.Options["botUsername"] = data.Telegram.BotUsername.ValueString()
			}
			if !data.Telegram.BotAPI.IsNull() {
				payload.Options["botAPI"] = data.Telegram.BotAPI.ValueString()
			}
			if !data.Telegram.ChatId.IsNull() {
				payload.Options["chatId"] = data.Telegram.ChatId.ValueString()
			}
			if !data.Telegram.SendSilently.IsNull() {
				payload.Options["sendSilently"] = data.Telegram.SendSilently.ValueBool()
			}
		}
	case "pushbullet":
		if data.Pushbullet != nil {
			if !data.Pushbullet.AccessToken.IsNull() {
				payload.Options["accessToken"] = data.Pushbullet.AccessToken.ValueString()
			}
			if !data.Pushbullet.ChannelTag.IsNull() {
				payload.Options["channelTag"] = data.Pushbullet.ChannelTag.ValueString()
			}
		}
	case "pushover":
		if data.Pushover != nil {
			if !data.Pushover.AccessToken.IsNull() {
				payload.Options["accessToken"] = data.Pushover.AccessToken.ValueString()
			}
			if !data.Pushover.UserToken.IsNull() {
				payload.Options["userToken"] = data.Pushover.UserToken.ValueString()
			}
			if !data.Pushover.Sound.IsNull() {
				payload.Options["sound"] = data.Pushover.Sound.ValueString()
			}
		}
	case "ntfy":
		if data.Ntfy != nil {
			if !data.Ntfy.Url.IsNull() {
				payload.Options["url"] = data.Ntfy.Url.ValueString()
			}
			if !data.Ntfy.Topic.IsNull() {
				payload.Options["topic"] = data.Ntfy.Topic.ValueString()
			}
			if !data.Ntfy.AuthMethodUsernamePassword.IsNull() {
				payload.Options["authMethodUsernamePassword"] = data.Ntfy.AuthMethodUsernamePassword.ValueBool()
			}
			if !data.Ntfy.Username.IsNull() {
				payload.Options["username"] = data.Ntfy.Username.ValueString()
			}
			if !data.Ntfy.Password.IsNull() {
				payload.Options["password"] = data.Ntfy.Password.ValueString()
			}
			if !data.Ntfy.AuthMethodToken.IsNull() {
				payload.Options["authMethodToken"] = data.Ntfy.AuthMethodToken.ValueBool()
			}
			if !data.Ntfy.Token.IsNull() {
				payload.Options["token"] = data.Ntfy.Token.ValueString()
			}
			if !data.Ntfy.Priority.IsNull() {
				payload.Options["priority"] = data.Ntfy.Priority.ValueInt64()
			}
		}
	case "webhook":
		if data.Webhook != nil {
			if !data.Webhook.WebhookUrl.IsNull() {
				payload.Options["webhookUrl"] = data.Webhook.WebhookUrl.ValueString()
			}
			if !data.Webhook.JsonPayload.IsNull() {
				payload.Options["jsonPayload"] = data.Webhook.JsonPayload.ValueString()
			}
			if !data.Webhook.AuthHeader.IsNull() {
				payload.Options["authHeader"] = data.Webhook.AuthHeader.ValueString()
			}
		}
	case "gotify":
		if data.Gotify != nil {
			if !data.Gotify.Url.IsNull() {
				payload.Options["url"] = data.Gotify.Url.ValueString()
			}
			if !data.Gotify.Token.IsNull() {
				payload.Options["token"] = data.Gotify.Token.ValueString()
			}
		}
	case "webpush":
		// no options
	default:
		return "", fmt.Errorf("unsupported agent: %s", data.Agent.ValueString())
	}

	b, err := json.Marshal(payload)
	return string(b), err
}

func parsePayload(data *NotificationAgentModel, body []byte) error {
	var payload notificationAgentPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	data.Enabled = types.BoolValue(payload.Enabled)
	data.EmbedPoster = types.BoolValue(payload.EmbedPoster)
	data.TypesMask = types.Int64Value(payload.Types)

	mask := payload.Types
	data.OnRequestPending = types.BoolValue(mask&1 != 0)
	data.OnRequestApproved = types.BoolValue(mask&2 != 0)
	data.OnRequestRejected = types.BoolValue(mask&4 != 0)
	data.OnRequestFailed = types.BoolValue(mask&8 != 0)
	data.OnRequestAvailable = types.BoolValue(mask&16 != 0)
	data.OnRequestDeclined = types.BoolValue(mask&32 != 0)
	data.OnRequestAutoApproved = types.BoolValue(mask&64 != 0)
	data.OnMediaAvailable = types.BoolValue(mask&128 != 0)
	data.OnMediaFailed = types.BoolValue(mask&256 != 0)
	data.OnMediaSkipped = types.BoolValue(mask&512 != 0)
	data.OnMediaIssued = types.BoolValue(mask&1024 != 0)
	data.OnMediaFollowed = types.BoolValue(mask&2048 != 0)

	opt := payload.Options
	getString := func(key string) types.String {
		if v, ok := opt[key].(string); ok {
			return types.StringValue(v)
		}
		return types.StringNull()
	}
	getBool := func(key string) types.Bool {
		if v, ok := opt[key].(bool); ok {
			return types.BoolValue(v)
		}
		return types.BoolNull()
	}
	getInt64 := func(key string) types.Int64 {
		if v, ok := opt[key].(float64); ok {
			return types.Int64Value(int64(v))
		}
		if v, ok := opt[key].(int64); ok {
			return types.Int64Value(v)
		}
		if v, ok := opt[key].(int); ok {
			return types.Int64Value(int64(v))
		}
		return types.Int64Null()
	}

	// Reset blocks to nil initially
	data.Discord = nil
	data.Slack = nil
	data.Email = nil
	data.LunaSea = nil
	data.Telegram = nil
	data.Pushbullet = nil
	data.Pushover = nil
	data.Ntfy = nil
	data.Webhook = nil
	data.Gotify = nil
	data.Webpush = nil

	switch data.Agent.ValueString() {
	case "discord":
		data.Discord = &NotificationAgentDiscordModel{
			BotUsername:    getString("botUsername"),
			BotAvatarUrl:   getString("botAvatarUrl"),
			WebhookUrl:     getString("webhookUrl"),
			EnableMentions: getBool("enableMentions"),
		}
	case "slack":
		data.Slack = &NotificationAgentSlackModel{
			WebhookUrl: getString("webhookUrl"),
		}
	case "email":
		data.Email = &NotificationAgentEmailModel{
			EmailFrom:       getString("emailFrom"),
			SmtpHost:        getString("smtpHost"),
			SmtpPort:        getInt64("smtpPort"),
			Secure:          getBool("secure"),
			IgnoreTls:       getBool("ignoreTls"),
			RequireTls:      getBool("requireTls"),
			AuthUser:        getString("authUser"),
			AuthPass:        getString("authPass"),
			AllowSelfSigned: getBool("allowSelfSigned"),
			SenderName:      getString("senderName"),
			PgpPrivateKey:   getString("pgpPrivateKey"),
			PgpPassword:     getString("pgpPassword"),
		}
	case "lunasea":
		data.LunaSea = &NotificationAgentLunaSeaModel{
			WebhookUrl:  getString("webhookUrl"),
			ProfileName: getString("profileName"),
		}
	case "telegram":
		data.Telegram = &NotificationAgentTelegramModel{
			BotUsername:  getString("botUsername"),
			BotAPI:       getString("botAPI"),
			ChatId:       getString("chatId"),
			SendSilently: getBool("sendSilently"),
		}
	case "pushbullet":
		data.Pushbullet = &NotificationAgentPushbulletModel{
			AccessToken: getString("accessToken"),
			ChannelTag:  getString("channelTag"),
		}
	case "pushover":
		data.Pushover = &NotificationAgentPushoverModel{
			AccessToken: getString("accessToken"),
			UserToken:   getString("userToken"),
			Sound:       getString("sound"),
		}
	case "ntfy":
		data.Ntfy = &NotificationAgentNtfyModel{
			Url:                        getString("url"),
			Topic:                      getString("topic"),
			AuthMethodUsernamePassword: getBool("authMethodUsernamePassword"),
			Username:                   getString("username"),
			Password:                   getString("password"),
			AuthMethodToken:            getBool("authMethodToken"),
			Token:                      getString("token"),
			Priority:                   getInt64("priority"),
		}
	case "webhook":
		data.Webhook = &NotificationAgentWebhookModel{
			WebhookUrl:  getString("webhookUrl"),
			JsonPayload: getString("jsonPayload"),
			AuthHeader:  getString("authHeader"),
		}
	case "gotify":
		data.Gotify = &NotificationAgentGotifyModel{
			Url:   getString("url"),
			Token: getString("token"),
		}
	case "webpush":
		data.Webpush = &NotificationAgentWebpushModel{}
	}

	return nil
}

func (r *NotificationAgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NotificationAgentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := notificationPath(data.Agent.ValueString())
	payloadStr, err := buildPayload(&data)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	res, err := r.client.Request(ctx, "POST", path, payloadStr, nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue(data.Agent.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationAgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NotificationAgentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := notificationPath(data.Agent.ValueString())
	res, err := r.client.Request(ctx, "GET", path, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if res.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	if err := parsePayload(&data, res.Body); err != nil {
		resp.Diagnostics.AddError("Parse Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationAgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NotificationAgentModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := notificationPath(data.Agent.ValueString())
	payloadStr, err := buildPayload(&data)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	res, err := r.client.Request(ctx, "POST", path, payloadStr, nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}
	data.ID = types.StringValue(data.Agent.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NotificationAgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NotificationAgentModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := notificationPath(data.Agent.ValueString())

	// Default disable payload upon deletion to clear it
	disablePayload := `{"enabled":false,"types":0,"options":{}}`
	res, err := r.client.Request(ctx, "POST", path, disablePayload, nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		if !strings.Contains(string(res.Body), "Unknown notification agent") && res.StatusCode != 404 {
			resp.Diagnostics.AddWarning("Delete Error", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		}
	}
}

func (r *NotificationAgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("agent"), req.ID)...)
}
