package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

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
	OnIssueCreated        types.Bool                        `tfsdk:"on_issue_created"`
	OnIssueComment        types.Bool                        `tfsdk:"on_issue_comment"`
	OnIssueResolved       types.Bool                        `tfsdk:"on_issue_resolved"`
	OnIssueReopened       types.Bool                        `tfsdk:"on_issue_reopened"`
	OnMediaAutoRequested  types.Bool                        `tfsdk:"on_media_auto_requested"`
}

type notificationAgentPayload struct {
	Enabled     bool                   `json:"enabled"`
	EmbedPoster bool                   `json:"embedPoster"`
	Types       int64                  `json:"types"`
	Options     map[string]interface{} `json:"options"`
}

type notificationClientCommonModel struct {
	ID          types.String `tfsdk:"id"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	EmbedPoster types.Bool   `tfsdk:"embed_poster"`
	TypesMask   types.Int64  `tfsdk:"types"`

	OnRequestPending      types.Bool `tfsdk:"on_request_pending"`
	OnRequestApproved     types.Bool `tfsdk:"on_request_approved"`
	OnRequestRejected     types.Bool `tfsdk:"on_request_rejected"`
	OnRequestFailed       types.Bool `tfsdk:"on_request_failed"`
	OnRequestAvailable    types.Bool `tfsdk:"on_request_available"`
	OnRequestDeclined     types.Bool `tfsdk:"on_request_declined"`
	OnRequestAutoApproved types.Bool `tfsdk:"on_request_auto_approved"`
	OnMediaAvailable      types.Bool `tfsdk:"on_media_available"`
	OnMediaFailed         types.Bool `tfsdk:"on_media_failed"`
	OnMediaSkipped        types.Bool `tfsdk:"on_media_skipped"`
	OnMediaIssued         types.Bool `tfsdk:"on_media_issued"`
	OnMediaFollowed       types.Bool `tfsdk:"on_media_followed"`
	OnIssueCreated        types.Bool `tfsdk:"on_issue_created"`
	OnIssueComment        types.Bool `tfsdk:"on_issue_comment"`
	OnIssueResolved       types.Bool `tfsdk:"on_issue_resolved"`
	OnIssueReopened       types.Bool `tfsdk:"on_issue_reopened"`
	OnMediaAutoRequested  types.Bool `tfsdk:"on_media_auto_requested"`
}

type notificationAttributeReader interface {
	GetAttribute(context.Context, path.Path, any) diag.Diagnostics
}

type notificationAttributeWriter interface {
	SetAttribute(context.Context, path.Path, any) diag.Diagnostics
}

type NotificationClientResource struct {
	client *APIClient
	agent  string
}

var _ resource.Resource = &NotificationClientResource{}
var _ resource.ResourceWithImportState = &NotificationClientResource{}

func newNotificationClientResource(agent string) resource.Resource {
	return &NotificationClientResource{agent: agent}
}

func NewNotificationDiscordResource() resource.Resource {
	return newNotificationClientResource("discord")
}
func NewNotificationSlackResource() resource.Resource { return newNotificationClientResource("slack") }
func NewNotificationEmailResource() resource.Resource { return newNotificationClientResource("email") }
func NewNotificationLunaSeaResource() resource.Resource {
	return newNotificationClientResource("lunasea")
}
func NewNotificationTelegramResource() resource.Resource {
	return newNotificationClientResource("telegram")
}
func NewNotificationPushbulletResource() resource.Resource {
	return newNotificationClientResource("pushbullet")
}
func NewNotificationPushoverResource() resource.Resource {
	return newNotificationClientResource("pushover")
}
func NewNotificationNtfyResource() resource.Resource { return newNotificationClientResource("ntfy") }
func NewNotificationWebhookResource() resource.Resource {
	return newNotificationClientResource("webhook")
}
func NewNotificationGotifyResource() resource.Resource {
	return newNotificationClientResource("gotify")
}
func NewNotificationWebpushResource() resource.Resource {
	return newNotificationClientResource("webpush")
}

func (r *NotificationClientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_" + r.agent
}

func (r *NotificationClientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.UseStateForUnknown(),
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
	for name, attr := range notificationAgentResourceEventAttributes() {
		attributes[name] = attr
	}

	optionAttr, ok := notificationAgentResourceOptionAttribute(r.agent)
	if !ok {
		resp.Diagnostics.AddError("Unsupported notification agent", fmt.Sprintf("agent %q is not supported", r.agent))
		return
	}
	attributes[r.agent] = optionAttr

	resp.Schema = schema.Schema{
		MarkdownDescription: fmt.Sprintf("Manage Seerr %s notification settings via /api/v1/settings/notifications/%s.", r.agent, r.agent),
		Attributes:          attributes,
	}
}

func (r *NotificationClientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func applyCommonNotificationFields(data *NotificationAgentModel, common notificationClientCommonModel) {
	data.ID = common.ID
	data.Enabled = common.Enabled
	data.EmbedPoster = common.EmbedPoster
	data.TypesMask = common.TypesMask
	data.OnRequestPending = common.OnRequestPending
	data.OnRequestApproved = common.OnRequestApproved
	data.OnRequestRejected = common.OnRequestRejected
	data.OnRequestFailed = common.OnRequestFailed
	data.OnRequestAvailable = common.OnRequestAvailable
	data.OnRequestDeclined = common.OnRequestDeclined
	data.OnRequestAutoApproved = common.OnRequestAutoApproved
	data.OnMediaAvailable = common.OnMediaAvailable
	data.OnMediaFailed = common.OnMediaFailed
	data.OnMediaSkipped = common.OnMediaSkipped
	data.OnMediaIssued = common.OnMediaIssued
	data.OnMediaFollowed = common.OnMediaFollowed
	data.OnIssueCreated = common.OnIssueCreated
	data.OnIssueComment = common.OnIssueComment
	data.OnIssueResolved = common.OnIssueResolved
	data.OnIssueReopened = common.OnIssueReopened
	data.OnMediaAutoRequested = common.OnMediaAutoRequested
}

func commonNotificationFields(data *NotificationAgentModel) notificationClientCommonModel {
	return notificationClientCommonModel{
		ID:                    data.ID,
		Enabled:               data.Enabled,
		EmbedPoster:           data.EmbedPoster,
		TypesMask:             data.TypesMask,
		OnRequestPending:      data.OnRequestPending,
		OnRequestApproved:     data.OnRequestApproved,
		OnRequestRejected:     data.OnRequestRejected,
		OnRequestFailed:       data.OnRequestFailed,
		OnRequestAvailable:    data.OnRequestAvailable,
		OnRequestDeclined:     data.OnRequestDeclined,
		OnRequestAutoApproved: data.OnRequestAutoApproved,
		OnMediaAvailable:      data.OnMediaAvailable,
		OnMediaFailed:         data.OnMediaFailed,
		OnMediaSkipped:        data.OnMediaSkipped,
		OnMediaIssued:         data.OnMediaIssued,
		OnMediaFollowed:       data.OnMediaFollowed,
		OnIssueCreated:        data.OnIssueCreated,
		OnIssueComment:        data.OnIssueComment,
		OnIssueResolved:       data.OnIssueResolved,
		OnIssueReopened:       data.OnIssueReopened,
		OnMediaAutoRequested:  data.OnMediaAutoRequested,
	}
}

func readNotificationClientModel(ctx context.Context, reader notificationAttributeReader, agent string) (NotificationAgentModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	var common notificationClientCommonModel

	for _, field := range []struct {
		name   string
		target any
	}{
		{name: "id", target: &common.ID},
		{name: "enabled", target: &common.Enabled},
		{name: "embed_poster", target: &common.EmbedPoster},
		{name: "types", target: &common.TypesMask},
		{name: "on_request_pending", target: &common.OnRequestPending},
		{name: "on_request_approved", target: &common.OnRequestApproved},
		{name: "on_request_rejected", target: &common.OnRequestRejected},
		{name: "on_request_failed", target: &common.OnRequestFailed},
		{name: "on_request_available", target: &common.OnRequestAvailable},
		{name: "on_request_declined", target: &common.OnRequestDeclined},
		{name: "on_request_auto_approved", target: &common.OnRequestAutoApproved},
		{name: "on_media_available", target: &common.OnMediaAvailable},
		{name: "on_media_failed", target: &common.OnMediaFailed},
		{name: "on_media_skipped", target: &common.OnMediaSkipped},
		{name: "on_media_issued", target: &common.OnMediaIssued},
		{name: "on_media_followed", target: &common.OnMediaFollowed},
		{name: "on_issue_created", target: &common.OnIssueCreated},
		{name: "on_issue_comment", target: &common.OnIssueComment},
		{name: "on_issue_resolved", target: &common.OnIssueResolved},
		{name: "on_issue_reopened", target: &common.OnIssueReopened},
		{name: "on_media_auto_requested", target: &common.OnMediaAutoRequested},
	} {
		diags.Append(reader.GetAttribute(ctx, path.Root(field.name), field.target)...)
	}

	data := NotificationAgentModel{Agent: types.StringValue(agent)}
	applyCommonNotificationFields(&data, common)

	switch agent {
	case "discord":
		diags.Append(reader.GetAttribute(ctx, path.Root("discord"), &data.Discord)...)
	case "slack":
		diags.Append(reader.GetAttribute(ctx, path.Root("slack"), &data.Slack)...)
	case "email":
		diags.Append(reader.GetAttribute(ctx, path.Root("email"), &data.Email)...)
	case "lunasea":
		diags.Append(reader.GetAttribute(ctx, path.Root("lunasea"), &data.LunaSea)...)
	case "telegram":
		diags.Append(reader.GetAttribute(ctx, path.Root("telegram"), &data.Telegram)...)
	case "pushbullet":
		diags.Append(reader.GetAttribute(ctx, path.Root("pushbullet"), &data.Pushbullet)...)
	case "pushover":
		diags.Append(reader.GetAttribute(ctx, path.Root("pushover"), &data.Pushover)...)
	case "ntfy":
		diags.Append(reader.GetAttribute(ctx, path.Root("ntfy"), &data.Ntfy)...)
	case "webhook":
		diags.Append(reader.GetAttribute(ctx, path.Root("webhook"), &data.Webhook)...)
	case "gotify":
		diags.Append(reader.GetAttribute(ctx, path.Root("gotify"), &data.Gotify)...)
	case "webpush":
		diags.Append(reader.GetAttribute(ctx, path.Root("webpush"), &data.Webpush)...)
	default:
		diags.AddError("Unsupported notification agent", fmt.Sprintf("agent %q is not supported", agent))
	}

	return data, diags
}

func setNotificationClientState(ctx context.Context, writer notificationAttributeWriter, data *NotificationAgentModel) diag.Diagnostics {
	var diags diag.Diagnostics
	common := commonNotificationFields(data)

	for _, field := range []struct {
		name  string
		value any
	}{
		{name: "id", value: common.ID},
		{name: "enabled", value: common.Enabled},
		{name: "embed_poster", value: common.EmbedPoster},
		{name: "types", value: common.TypesMask},
		{name: "on_request_pending", value: common.OnRequestPending},
		{name: "on_request_approved", value: common.OnRequestApproved},
		{name: "on_request_rejected", value: common.OnRequestRejected},
		{name: "on_request_failed", value: common.OnRequestFailed},
		{name: "on_request_available", value: common.OnRequestAvailable},
		{name: "on_request_declined", value: common.OnRequestDeclined},
		{name: "on_request_auto_approved", value: common.OnRequestAutoApproved},
		{name: "on_media_available", value: common.OnMediaAvailable},
		{name: "on_media_failed", value: common.OnMediaFailed},
		{name: "on_media_skipped", value: common.OnMediaSkipped},
		{name: "on_media_issued", value: common.OnMediaIssued},
		{name: "on_media_followed", value: common.OnMediaFollowed},
		{name: "on_issue_created", value: common.OnIssueCreated},
		{name: "on_issue_comment", value: common.OnIssueComment},
		{name: "on_issue_resolved", value: common.OnIssueResolved},
		{name: "on_issue_reopened", value: common.OnIssueReopened},
		{name: "on_media_auto_requested", value: common.OnMediaAutoRequested},
	} {
		diags.Append(writer.SetAttribute(ctx, path.Root(field.name), field.value)...)
	}

	switch data.Agent.ValueString() {
	case "discord":
		diags.Append(writer.SetAttribute(ctx, path.Root("discord"), data.Discord)...)
	case "slack":
		diags.Append(writer.SetAttribute(ctx, path.Root("slack"), data.Slack)...)
	case "email":
		diags.Append(writer.SetAttribute(ctx, path.Root("email"), data.Email)...)
	case "lunasea":
		diags.Append(writer.SetAttribute(ctx, path.Root("lunasea"), data.LunaSea)...)
	case "telegram":
		diags.Append(writer.SetAttribute(ctx, path.Root("telegram"), data.Telegram)...)
	case "pushbullet":
		diags.Append(writer.SetAttribute(ctx, path.Root("pushbullet"), data.Pushbullet)...)
	case "pushover":
		diags.Append(writer.SetAttribute(ctx, path.Root("pushover"), data.Pushover)...)
	case "ntfy":
		diags.Append(writer.SetAttribute(ctx, path.Root("ntfy"), data.Ntfy)...)
	case "webhook":
		diags.Append(writer.SetAttribute(ctx, path.Root("webhook"), data.Webhook)...)
	case "gotify":
		diags.Append(writer.SetAttribute(ctx, path.Root("gotify"), data.Gotify)...)
	case "webpush":
		diags.Append(writer.SetAttribute(ctx, path.Root("webpush"), data.Webpush)...)
	default:
		diags.AddError("Unsupported notification agent", fmt.Sprintf("agent %q is not supported", data.Agent.ValueString()))
	}

	return diags
}

func buildPayload(data *NotificationAgentModel) (string, error) {
	payload := notificationAgentPayload{
		Enabled:     data.Enabled.ValueBool(),
		EmbedPoster: data.EmbedPoster.ValueBool(),
		Types:       data.TypesMask.ValueInt64(),
		Options:     make(map[string]interface{}),
	}

	mask := data.TypesMask.ValueInt64()
	updateMask := func(val types.Bool, bit int64) {
		if !val.IsNull() && !val.IsUnknown() {
			if val.ValueBool() {
				mask |= bit
			} else {
				mask &= ^bit
			}
		}
	}

	updateMask(data.OnRequestPending, 2)
	updateMask(data.OnRequestApproved, 4)
	updateMask(data.OnMediaAvailable, 8)
	updateMask(data.OnMediaFailed, 16)
	updateMask(data.OnRequestDeclined, 64)
	updateMask(data.OnRequestAutoApproved, 128)
	updateMask(data.OnIssueCreated, 256)
	updateMask(data.OnIssueComment, 512)
	updateMask(data.OnIssueResolved, 1024)
	updateMask(data.OnIssueReopened, 2048)
	updateMask(data.OnMediaAutoRequested, 4096)

	updateMask(data.OnRequestRejected, 4)
	updateMask(data.OnRequestFailed, 8)
	updateMask(data.OnRequestAvailable, 16)
	updateMask(data.OnMediaSkipped, 512)
	updateMask(data.OnMediaIssued, 1024)
	updateMask(data.OnMediaFollowed, 2048)

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
	data.OnRequestPending = types.BoolValue(mask&2 != 0)
	data.OnRequestApproved = types.BoolValue(mask&4 != 0)
	data.OnRequestRejected = types.BoolValue(mask&4 != 0)
	data.OnRequestFailed = types.BoolValue(mask&8 != 0)
	data.OnRequestAvailable = types.BoolValue(mask&16 != 0)
	data.OnRequestDeclined = types.BoolValue(mask&64 != 0)
	data.OnRequestAutoApproved = types.BoolValue(mask&128 != 0)
	data.OnMediaAvailable = types.BoolValue(mask&8 != 0)
	data.OnMediaFailed = types.BoolValue(mask&16 != 0)
	data.OnMediaSkipped = types.BoolValue(mask&512 != 0)
	data.OnMediaIssued = types.BoolValue(mask&1024 != 0)
	data.OnMediaFollowed = types.BoolValue(mask&2048 != 0)
	data.OnIssueCreated = types.BoolValue(mask&256 != 0)
	data.OnIssueComment = types.BoolValue(mask&512 != 0)
	data.OnIssueResolved = types.BoolValue(mask&1024 != 0)
	data.OnIssueReopened = types.BoolValue(mask&2048 != 0)
	data.OnMediaAutoRequested = types.BoolValue(mask&4096 != 0)

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
		data.Slack = &NotificationAgentSlackModel{WebhookUrl: getString("webhookUrl")}
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
		data.LunaSea = &NotificationAgentLunaSeaModel{WebhookUrl: getString("webhookUrl"), ProfileName: getString("profileName")}
	case "telegram":
		data.Telegram = &NotificationAgentTelegramModel{
			BotUsername:  getString("botUsername"),
			BotAPI:       getString("botAPI"),
			ChatId:       getString("chatId"),
			SendSilently: getBool("sendSilently"),
		}
	case "pushbullet":
		data.Pushbullet = &NotificationAgentPushbulletModel{AccessToken: getString("accessToken"), ChannelTag: getString("channelTag")}
	case "pushover":
		data.Pushover = &NotificationAgentPushoverModel{AccessToken: getString("accessToken"), UserToken: getString("userToken"), Sound: getString("sound")}
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
		data.Webhook = &NotificationAgentWebhookModel{WebhookUrl: getString("webhookUrl"), JsonPayload: getString("jsonPayload"), AuthHeader: getString("authHeader")}
	case "gotify":
		data.Gotify = &NotificationAgentGotifyModel{Url: getString("url"), Token: getString("token")}
	case "webpush":
		data.Webpush = &NotificationAgentWebpushModel{}
	}

	return nil
}

func (r *NotificationClientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	data, diags := readNotificationClientModel(ctx, req.Plan, r.agent)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payloadStr, err := buildPayload(&data)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	res, err := r.client.Request(ctx, "POST", notificationPath(r.agent), payloadStr, nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	// Capture plan state to preserve sensitive fields
	planData := data

	if err := parsePayload(&data, res.Body); err != nil {
		resp.Diagnostics.AddError("Parse Failed", err.Error())
		return
	}

	// Preserve sensitive fields from plan if API didn't return them
	preserveSensitiveNotificationFields(&data, &planData)

	data.ID = types.StringValue(r.agent)
	resp.Diagnostics.Append(setNotificationClientState(ctx, &resp.State, &data)...)
}

func (r *NotificationClientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	data := NotificationAgentModel{Agent: types.StringValue(r.agent)}

	res, err := r.client.Request(ctx, "GET", notificationPath(r.agent), "", nil)
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

	// Capture current state to preserve sensitive fields
	var state NotificationAgentModel
	diags := req.State.Get(ctx, &state)
	if !diags.HasError() {
		preserveSensitiveNotificationFields(&data, &state)
	}

	data.ID = types.StringValue(r.agent)
	resp.Diagnostics.Append(setNotificationClientState(ctx, &resp.State, &data)...)
}

func (r *NotificationClientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	data, diags := readNotificationClientModel(ctx, req.Plan, r.agent)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	payloadStr, err := buildPayload(&data)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	res, err := r.client.Request(ctx, "POST", notificationPath(r.agent), payloadStr, nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	// Capture plan state to preserve sensitive fields
	planData := data

	if err := parsePayload(&data, res.Body); err != nil {
		resp.Diagnostics.AddError("Parse Failed", err.Error())
		return
	}

	// Preserve sensitive fields from plan if API didn't return them
	preserveSensitiveNotificationFields(&data, &planData)

	data.ID = types.StringValue(r.agent)
	resp.Diagnostics.Append(setNotificationClientState(ctx, &resp.State, &data)...)
}

func preserveSensitiveNotificationFields(data, source *NotificationAgentModel) {
	if source == nil {
		return
	}

	switch data.Agent.ValueString() {
	case "email":
		if data.Email != nil && source.Email != nil {
			if data.Email.AuthPass.IsNull() {
				data.Email.AuthPass = source.Email.AuthPass
			}
			if data.Email.PgpPrivateKey.IsNull() {
				data.Email.PgpPrivateKey = source.Email.PgpPrivateKey
			}
			if data.Email.PgpPassword.IsNull() {
				data.Email.PgpPassword = source.Email.PgpPassword
			}
		}
	case "telegram":
		if data.Telegram != nil && source.Telegram != nil {
			if data.Telegram.BotAPI.IsNull() {
				data.Telegram.BotAPI = source.Telegram.BotAPI
			}
		}
	case "pushbullet":
		if data.Pushbullet != nil && source.Pushbullet != nil {
			if data.Pushbullet.AccessToken.IsNull() {
				data.Pushbullet.AccessToken = source.Pushbullet.AccessToken
			}
		}
	case "pushover":
		if data.Pushover != nil && source.Pushover != nil {
			if data.Pushover.AccessToken.IsNull() {
				data.Pushover.AccessToken = source.Pushover.AccessToken
			}
			if data.Pushover.UserToken.IsNull() {
				data.Pushover.UserToken = source.Pushover.UserToken
			}
		}
	case "ntfy":
		if data.Ntfy != nil && source.Ntfy != nil {
			if data.Ntfy.Password.IsNull() {
				data.Ntfy.Password = source.Ntfy.Password
			}
			if data.Ntfy.Token.IsNull() {
				data.Ntfy.Token = source.Ntfy.Token
			}
		}
	case "webhook":
		if data.Webhook != nil && source.Webhook != nil {
			if data.Webhook.AuthHeader.IsNull() {
				data.Webhook.AuthHeader = source.Webhook.AuthHeader
			}
		}
	case "gotify":
		if data.Gotify != nil && source.Gotify != nil {
			if data.Gotify.Token.IsNull() {
				data.Gotify.Token = source.Gotify.Token
			}
		}
	}
}

func (r *NotificationClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	disablePayload := `{"enabled":false,"types":0,"options":{}}`
	res, err := r.client.Request(ctx, "POST", notificationPath(r.agent), disablePayload, nil)
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

func (r *NotificationClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != r.agent {
		resp.Diagnostics.AddError("Invalid import id", fmt.Sprintf("use import id %q for this resource type", r.agent))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
