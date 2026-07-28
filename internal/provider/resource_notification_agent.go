package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationAgentModel struct {
	ID          types.String `tfsdk:"id"`
	Agent       types.String `tfsdk:"agent"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	EmbedPoster types.Bool   `tfsdk:"embed_poster"`

	Discord           *NotificationAgentDiscordModel    `tfsdk:"discord"`
	Slack             *NotificationAgentSlackModel      `tfsdk:"slack"`
	Email             *NotificationAgentEmailModel      `tfsdk:"email"`
	LunaSea           *NotificationAgentLunaSeaModel    `tfsdk:"lunasea"`
	Telegram          *NotificationAgentTelegramModel   `tfsdk:"telegram"`
	Pushbullet        *NotificationAgentPushbulletModel `tfsdk:"pushbullet"`
	Pushover          *NotificationAgentPushoverModel   `tfsdk:"pushover"`
	Ntfy              *NotificationAgentNtfyModel       `tfsdk:"ntfy"`
	Webhook           *NotificationAgentWebhookModel    `tfsdk:"webhook"`
	Gotify            *NotificationAgentGotifyModel     `tfsdk:"gotify"`
	Webpush           *NotificationAgentWebpushModel    `tfsdk:"webpush"`
	NotificationTypes types.Set                         `tfsdk:"notification_types"`
}

type notificationAgentPayload struct {
	Enabled     bool                   `json:"enabled"`
	EmbedPoster bool                   `json:"embedPoster"`
	Types       int64                  `json:"types"`
	Options     map[string]interface{} `json:"options"`
}

type notificationClientCommonModel struct {
	ID                types.String `tfsdk:"id"`
	Enabled           types.Bool   `tfsdk:"enabled"`
	EmbedPoster       types.Bool   `tfsdk:"embed_poster"`
	NotificationTypes types.Set    `tfsdk:"notification_types"`
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
	data.NotificationTypes = common.NotificationTypes
}

func commonNotificationFields(data *NotificationAgentModel) notificationClientCommonModel {
	return notificationClientCommonModel{
		ID:                data.ID,
		Enabled:           data.Enabled,
		EmbedPoster:       data.EmbedPoster,
		NotificationTypes: data.NotificationTypes,
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
		{name: "notification_types", target: &common.NotificationTypes},
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
		{name: "notification_types", value: common.NotificationTypes},
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

func isKnownNonNullString(v types.String) bool {
	return !v.IsNull() && !v.IsUnknown()
}

func isKnownNonNullBool(v types.Bool) bool {
	return !v.IsNull() && !v.IsUnknown()
}

func isKnownNonNullInt64(v types.Int64) bool {
	return !v.IsNull() && !v.IsUnknown()
}

func buildPayload(ctx context.Context, data *NotificationAgentModel) (string, error) {
	payload := notificationAgentPayload{
		Enabled:     data.Enabled.ValueBool(),
		EmbedPoster: data.EmbedPoster.ValueBool(),
		Types:       notificationTypesMask(ctx, data),
		Options:     make(map[string]interface{}),
	}

	switch data.Agent.ValueString() {
	case "discord":
		if data.Discord != nil {
			if isKnownNonNullString(data.Discord.BotUsername) {
				payload.Options["botUsername"] = data.Discord.BotUsername.ValueString()
			}
			if isKnownNonNullString(data.Discord.BotAvatarUrl) {
				payload.Options["botAvatarUrl"] = data.Discord.BotAvatarUrl.ValueString()
			}
			if isKnownNonNullString(data.Discord.WebhookUrl) {
				payload.Options["webhookUrl"] = data.Discord.WebhookUrl.ValueString()
			}
			if isKnownNonNullBool(data.Discord.EnableMentions) {
				payload.Options["enableMentions"] = data.Discord.EnableMentions.ValueBool()
			}
		}
	case "slack":
		if data.Slack != nil {
			if isKnownNonNullString(data.Slack.WebhookUrl) {
				payload.Options["webhookUrl"] = data.Slack.WebhookUrl.ValueString()
			}
		}
	case "email":
		if data.Email != nil {
			if isKnownNonNullString(data.Email.EmailFrom) {
				payload.Options["emailFrom"] = data.Email.EmailFrom.ValueString()
			}
			if isKnownNonNullString(data.Email.SmtpHost) {
				payload.Options["smtpHost"] = data.Email.SmtpHost.ValueString()
			}
			if isKnownNonNullInt64(data.Email.SmtpPort) {
				payload.Options["smtpPort"] = data.Email.SmtpPort.ValueInt64()
			}
			if isKnownNonNullBool(data.Email.Secure) {
				payload.Options["secure"] = data.Email.Secure.ValueBool()
			}
			if isKnownNonNullBool(data.Email.IgnoreTls) {
				payload.Options["ignoreTls"] = data.Email.IgnoreTls.ValueBool()
			}
			if isKnownNonNullBool(data.Email.RequireTls) {
				payload.Options["requireTls"] = data.Email.RequireTls.ValueBool()
			}
			if isKnownNonNullString(data.Email.AuthUser) {
				payload.Options["authUser"] = data.Email.AuthUser.ValueString()
			}
			if isKnownNonNullString(data.Email.AuthPass) {
				payload.Options["authPass"] = data.Email.AuthPass.ValueString()
			}
			if isKnownNonNullBool(data.Email.AllowSelfSigned) {
				payload.Options["allowSelfSigned"] = data.Email.AllowSelfSigned.ValueBool()
			}
			if isKnownNonNullString(data.Email.SenderName) {
				payload.Options["senderName"] = data.Email.SenderName.ValueString()
			}
			if isKnownNonNullString(data.Email.PgpPrivateKey) {
				payload.Options["pgpPrivateKey"] = data.Email.PgpPrivateKey.ValueString()
			}
			if isKnownNonNullString(data.Email.PgpPassword) {
				payload.Options["pgpPassword"] = data.Email.PgpPassword.ValueString()
			}
		}
	case "lunasea":
		if data.LunaSea != nil {
			if isKnownNonNullString(data.LunaSea.WebhookUrl) {
				payload.Options["webhookUrl"] = data.LunaSea.WebhookUrl.ValueString()
			}
			if isKnownNonNullString(data.LunaSea.ProfileName) {
				payload.Options["profileName"] = data.LunaSea.ProfileName.ValueString()
			}
		}
	case "telegram":
		if data.Telegram != nil {
			if isKnownNonNullString(data.Telegram.BotUsername) {
				payload.Options["botUsername"] = data.Telegram.BotUsername.ValueString()
			}
			if isKnownNonNullString(data.Telegram.BotAPI) {
				payload.Options["botAPI"] = data.Telegram.BotAPI.ValueString()
			}
			if isKnownNonNullString(data.Telegram.ChatId) {
				payload.Options["chatId"] = data.Telegram.ChatId.ValueString()
			}
			if isKnownNonNullBool(data.Telegram.SendSilently) {
				payload.Options["sendSilently"] = data.Telegram.SendSilently.ValueBool()
			}
		}
	case "pushbullet":
		if data.Pushbullet != nil {
			if isKnownNonNullString(data.Pushbullet.AccessToken) {
				payload.Options["accessToken"] = data.Pushbullet.AccessToken.ValueString()
			}
			if isKnownNonNullString(data.Pushbullet.ChannelTag) {
				payload.Options["channelTag"] = data.Pushbullet.ChannelTag.ValueString()
			}
		}
	case "pushover":
		if data.Pushover != nil {
			if isKnownNonNullString(data.Pushover.AccessToken) {
				payload.Options["accessToken"] = data.Pushover.AccessToken.ValueString()
			}
			if isKnownNonNullString(data.Pushover.UserToken) {
				payload.Options["userToken"] = data.Pushover.UserToken.ValueString()
			}
			if isKnownNonNullString(data.Pushover.Sound) {
				payload.Options["sound"] = data.Pushover.Sound.ValueString()
			}
		}
	case "ntfy":
		if data.Ntfy != nil {
			if isKnownNonNullString(data.Ntfy.Url) {
				payload.Options["url"] = data.Ntfy.Url.ValueString()
			}
			if isKnownNonNullString(data.Ntfy.Topic) {
				payload.Options["topic"] = data.Ntfy.Topic.ValueString()
			}
			if isKnownNonNullBool(data.Ntfy.AuthMethodUsernamePassword) {
				payload.Options["authMethodUsernamePassword"] = data.Ntfy.AuthMethodUsernamePassword.ValueBool()
			}
			if isKnownNonNullString(data.Ntfy.Username) {
				payload.Options["username"] = data.Ntfy.Username.ValueString()
			}
			if isKnownNonNullString(data.Ntfy.Password) {
				payload.Options["password"] = data.Ntfy.Password.ValueString()
			}
			if isKnownNonNullBool(data.Ntfy.AuthMethodToken) {
				payload.Options["authMethodToken"] = data.Ntfy.AuthMethodToken.ValueBool()
			}
			if isKnownNonNullString(data.Ntfy.Token) {
				payload.Options["token"] = data.Ntfy.Token.ValueString()
			}
			if isKnownNonNullInt64(data.Ntfy.Priority) {
				payload.Options["priority"] = data.Ntfy.Priority.ValueInt64()
			}
		}
	case "webhook":
		if data.Webhook != nil {
			if isKnownNonNullString(data.Webhook.WebhookUrl) {
				payload.Options["webhookUrl"] = data.Webhook.WebhookUrl.ValueString()
			}
			if isKnownNonNullString(data.Webhook.JsonPayload) {
				payload.Options["jsonPayload"] = data.Webhook.JsonPayload.ValueString()
			}
			if isKnownNonNullString(data.Webhook.AuthHeader) {
				payload.Options["authHeader"] = data.Webhook.AuthHeader.ValueString()
			}
		}
	case "gotify":
		if data.Gotify != nil {
			if isKnownNonNullString(data.Gotify.Url) {
				payload.Options["url"] = data.Gotify.Url.ValueString()
			}
			if isKnownNonNullString(data.Gotify.Token) {
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

func notificationTypesMask(ctx context.Context, data *NotificationAgentModel) int64 {
	if data.NotificationTypes.IsNull() || data.NotificationTypes.IsUnknown() {
		return 0
	}
	var typesList []string
	data.NotificationTypes.ElementsAs(ctx, &typesList, false)

	var mask int64 = 0
	for _, t := range typesList {
		switch t {
		case "MEDIA_PENDING":
			mask |= 2
		case "MEDIA_APPROVED":
			mask |= 4
		case "MEDIA_AVAILABLE":
			mask |= 8
		case "MEDIA_FAILED":
			mask |= 16
		case "MEDIA_DECLINED":
			mask |= 64
		case "MEDIA_AUTO_APPROVED":
			mask |= 128
		case "ISSUE_CREATED":
			mask |= 256
		case "ISSUE_COMMENT":
			mask |= 512
		case "ISSUE_RESOLVED":
			mask |= 1024
		case "ISSUE_REOPENED":
			mask |= 2048
		case "MEDIA_AUTO_REQUESTED":
			mask |= 4096
		}
	}

	return mask
}

func notificationEventTypeNames() []string {
	return []string{
		"MEDIA_PENDING",
		"MEDIA_APPROVED",
		"MEDIA_AVAILABLE",
		"MEDIA_FAILED",
		"MEDIA_DECLINED",
		"MEDIA_AUTO_APPROVED",
		"ISSUE_CREATED",
		"ISSUE_COMMENT",
		"ISSUE_RESOLVED",
		"ISSUE_REOPENED",
		"MEDIA_AUTO_REQUESTED",
	}
}

func notificationEventTypeValidator() validator.String {
	return stringvalidator.OneOf(notificationEventTypeNames()...)
}

func parsePayload(ctx context.Context, data *NotificationAgentModel, body []byte) error {
	var payload notificationAgentPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	data.Enabled = types.BoolValue(payload.Enabled)
	data.EmbedPoster = types.BoolValue(payload.EmbedPoster)

	mask := payload.Types
	var eventNames []string
	if mask&2 != 0 {
		eventNames = append(eventNames, "MEDIA_PENDING")
	}
	if mask&4 != 0 {
		eventNames = append(eventNames, "MEDIA_APPROVED")
	}
	if mask&8 != 0 {
		eventNames = append(eventNames, "MEDIA_AVAILABLE")
	}
	if mask&16 != 0 {
		eventNames = append(eventNames, "MEDIA_FAILED")
	}
	if mask&64 != 0 {
		eventNames = append(eventNames, "MEDIA_DECLINED")
	}
	if mask&128 != 0 {
		eventNames = append(eventNames, "MEDIA_AUTO_APPROVED")
	}
	if mask&256 != 0 {
		eventNames = append(eventNames, "ISSUE_CREATED")
	}
	if mask&512 != 0 {
		eventNames = append(eventNames, "ISSUE_COMMENT")
	}
	if mask&1024 != 0 {
		eventNames = append(eventNames, "ISSUE_RESOLVED")
	}
	if mask&2048 != 0 {
		eventNames = append(eventNames, "ISSUE_REOPENED")
	}
	if mask&4096 != 0 {
		eventNames = append(eventNames, "MEDIA_AUTO_REQUESTED")
	}

	setVal, diags := types.SetValueFrom(ctx, types.StringType, eventNames)
	if diags.HasError() {
		return fmt.Errorf("build notification_types set: %v", diags)
	}
	data.NotificationTypes = setVal

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

	payloadStr, err := buildPayload(ctx, &data)
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

	// Capture plan state to preserve sensitive and optional fields
	planData := data

	if err := parsePayload(ctx, &data, res.Body); err != nil {
		resp.Diagnostics.AddError("Parse Failed", err.Error())
		return
	}

	normalizeNotificationModel(&data, &planData)
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

	if err := parsePayload(ctx, &data, res.Body); err != nil {
		resp.Diagnostics.AddError("Parse Failed", err.Error())
		return
	}

	// Capture current state to preserve sensitive and optional fields
	var state NotificationAgentModel
	diags := req.State.Get(ctx, &state)
	if !diags.HasError() {
		normalizeNotificationModel(&data, &state)
	} else {
		normalizeNotificationModel(&data, nil)
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

	payloadStr, err := buildPayload(ctx, &data)
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

	// Capture plan state to preserve sensitive and optional fields
	planData := data

	if err := parsePayload(ctx, &data, res.Body); err != nil {
		resp.Diagnostics.AddError("Parse Failed", err.Error())
		return
	}

	normalizeNotificationModel(&data, &planData)
	data.ID = types.StringValue(r.agent)
	resp.Diagnostics.Append(setNotificationClientState(ctx, &resp.State, &data)...)
}

func normalizeStringField(target *types.String, sourceVal types.String) {
	if target.IsNull() || target.ValueString() == "" {
		if !sourceVal.IsNull() && sourceVal.ValueString() != "" {
			*target = sourceVal
		} else if !sourceVal.IsNull() && sourceVal.ValueString() == "" {
			*target = types.StringValue("")
		} else {
			*target = types.StringNull()
		}
	}
}

func normalizeBoolField(target *types.Bool, sourceVal types.Bool) {
	if target.IsNull() {
		if !sourceVal.IsNull() {
			*target = sourceVal
		} else {
			*target = types.BoolValue(false)
		}
	}
}

func normalizeInt64Field(target *types.Int64, sourceVal types.Int64) {
	if target.IsNull() || target.ValueInt64() == 0 {
		if !sourceVal.IsNull() && sourceVal.ValueInt64() != 0 {
			*target = sourceVal
		} else if !sourceVal.IsNull() && sourceVal.ValueInt64() == 0 {
			*target = types.Int64Value(0)
		} else {
			*target = types.Int64Null()
		}
	}
}

func normalizeSetField(target *types.Set, sourceVal types.Set) {
	if target.IsNull() || len(target.Elements()) == 0 {
		if !sourceVal.IsNull() && len(sourceVal.Elements()) > 0 {
			*target = sourceVal
		} else if !sourceVal.IsNull() && len(sourceVal.Elements()) == 0 {
			setVal, _ := types.SetValueFrom(context.Background(), types.StringType, []string{})
			*target = setVal
		} else {
			*target = types.SetNull(types.StringType)
		}
	}
}

func normalizeNotificationModel(data, source *NotificationAgentModel) {
	var srcTypes types.Set
	if source != nil {
		srcTypes = source.NotificationTypes
	}
	normalizeSetField(&data.NotificationTypes, srcTypes)

	switch data.Agent.ValueString() {
	case "discord":
		if data.Discord != nil {
			var srcBotUsername, srcBotAvatarUrl, srcWebhookUrl types.String
			var srcEnableMentions types.Bool
			if source != nil && source.Discord != nil {
				srcBotUsername = source.Discord.BotUsername
				srcBotAvatarUrl = source.Discord.BotAvatarUrl
				srcWebhookUrl = source.Discord.WebhookUrl
				srcEnableMentions = source.Discord.EnableMentions
			}
			normalizeStringField(&data.Discord.BotUsername, srcBotUsername)
			normalizeStringField(&data.Discord.BotAvatarUrl, srcBotAvatarUrl)
			normalizeStringField(&data.Discord.WebhookUrl, srcWebhookUrl)
			normalizeBoolField(&data.Discord.EnableMentions, srcEnableMentions)
		}
	case "slack":
		if data.Slack != nil {
			var srcWebhookUrl types.String
			if source != nil && source.Slack != nil {
				srcWebhookUrl = source.Slack.WebhookUrl
			}
			normalizeStringField(&data.Slack.WebhookUrl, srcWebhookUrl)
		}
	case "email":
		if data.Email != nil {
			var srcEmailFrom, srcSmtpHost, srcAuthUser, srcAuthPass, srcSenderName, srcPgpPrivateKey, srcPgpPassword types.String
			var srcSmtpPort types.Int64
			var srcSecure, srcIgnoreTls, srcRequireTls, srcAllowSelfSigned types.Bool
			if source != nil && source.Email != nil {
				srcEmailFrom = source.Email.EmailFrom
				srcSmtpHost = source.Email.SmtpHost
				srcSmtpPort = source.Email.SmtpPort
				srcSecure = source.Email.Secure
				srcIgnoreTls = source.Email.IgnoreTls
				srcRequireTls = source.Email.RequireTls
				srcAuthUser = source.Email.AuthUser
				srcAuthPass = source.Email.AuthPass
				srcAllowSelfSigned = source.Email.AllowSelfSigned
				srcSenderName = source.Email.SenderName
				srcPgpPrivateKey = source.Email.PgpPrivateKey
				srcPgpPassword = source.Email.PgpPassword
			}
			normalizeStringField(&data.Email.EmailFrom, srcEmailFrom)
			normalizeStringField(&data.Email.SmtpHost, srcSmtpHost)
			normalizeInt64Field(&data.Email.SmtpPort, srcSmtpPort)
			normalizeBoolField(&data.Email.Secure, srcSecure)
			normalizeBoolField(&data.Email.IgnoreTls, srcIgnoreTls)
			normalizeBoolField(&data.Email.RequireTls, srcRequireTls)
			normalizeStringField(&data.Email.AuthUser, srcAuthUser)
			normalizeStringField(&data.Email.AuthPass, srcAuthPass)
			normalizeBoolField(&data.Email.AllowSelfSigned, srcAllowSelfSigned)
			normalizeStringField(&data.Email.SenderName, srcSenderName)
			normalizeStringField(&data.Email.PgpPrivateKey, srcPgpPrivateKey)
			normalizeStringField(&data.Email.PgpPassword, srcPgpPassword)
		}
	case "lunasea":
		if data.LunaSea != nil {
			var srcWebhookUrl, srcProfileName types.String
			if source != nil && source.LunaSea != nil {
				srcWebhookUrl = source.LunaSea.WebhookUrl
				srcProfileName = source.LunaSea.ProfileName
			}
			normalizeStringField(&data.LunaSea.WebhookUrl, srcWebhookUrl)
			normalizeStringField(&data.LunaSea.ProfileName, srcProfileName)
		}
	case "telegram":
		if data.Telegram != nil {
			var srcBotUsername, srcBotAPI, srcChatId types.String
			var srcSendSilently types.Bool
			if source != nil && source.Telegram != nil {
				srcBotUsername = source.Telegram.BotUsername
				srcBotAPI = source.Telegram.BotAPI
				srcChatId = source.Telegram.ChatId
				srcSendSilently = source.Telegram.SendSilently
			}
			normalizeStringField(&data.Telegram.BotUsername, srcBotUsername)
			normalizeStringField(&data.Telegram.BotAPI, srcBotAPI)
			normalizeStringField(&data.Telegram.ChatId, srcChatId)
			normalizeBoolField(&data.Telegram.SendSilently, srcSendSilently)
		}
	case "pushbullet":
		if data.Pushbullet != nil {
			var srcAccessToken, srcChannelTag types.String
			if source != nil && source.Pushbullet != nil {
				srcAccessToken = source.Pushbullet.AccessToken
				srcChannelTag = source.Pushbullet.ChannelTag
			}
			normalizeStringField(&data.Pushbullet.AccessToken, srcAccessToken)
			normalizeStringField(&data.Pushbullet.ChannelTag, srcChannelTag)
		}
	case "pushover":
		if data.Pushover != nil {
			var srcAccessToken, srcUserToken, srcSound types.String
			if source != nil && source.Pushover != nil {
				srcAccessToken = source.Pushover.AccessToken
				srcUserToken = source.Pushover.UserToken
				srcSound = source.Pushover.Sound
			}
			normalizeStringField(&data.Pushover.AccessToken, srcAccessToken)
			normalizeStringField(&data.Pushover.UserToken, srcUserToken)
			normalizeStringField(&data.Pushover.Sound, srcSound)
		}
	case "ntfy":
		if data.Ntfy != nil {
			var srcUrl, srcTopic, srcUsername, srcPassword, srcToken types.String
			var srcAuthUserPass, srcAuthToken types.Bool
			var srcPriority types.Int64
			if source != nil && source.Ntfy != nil {
				srcUrl = source.Ntfy.Url
				srcTopic = source.Ntfy.Topic
				srcAuthUserPass = source.Ntfy.AuthMethodUsernamePassword
				srcUsername = source.Ntfy.Username
				srcPassword = source.Ntfy.Password
				srcAuthToken = source.Ntfy.AuthMethodToken
				srcToken = source.Ntfy.Token
				srcPriority = source.Ntfy.Priority
			}
			normalizeStringField(&data.Ntfy.Url, srcUrl)
			normalizeStringField(&data.Ntfy.Topic, srcTopic)
			normalizeBoolField(&data.Ntfy.AuthMethodUsernamePassword, srcAuthUserPass)
			normalizeStringField(&data.Ntfy.Username, srcUsername)
			normalizeStringField(&data.Ntfy.Password, srcPassword)
			normalizeBoolField(&data.Ntfy.AuthMethodToken, srcAuthToken)
			normalizeStringField(&data.Ntfy.Token, srcToken)
			normalizeInt64Field(&data.Ntfy.Priority, srcPriority)
		}
	case "webhook":
		if data.Webhook != nil {
			var srcWebhookUrl, srcJsonPayload, srcAuthHeader types.String
			if source != nil && source.Webhook != nil {
				srcWebhookUrl = source.Webhook.WebhookUrl
				srcJsonPayload = source.Webhook.JsonPayload
				srcAuthHeader = source.Webhook.AuthHeader
			}
			normalizeStringField(&data.Webhook.WebhookUrl, srcWebhookUrl)
			normalizeStringField(&data.Webhook.JsonPayload, srcJsonPayload)
			normalizeStringField(&data.Webhook.AuthHeader, srcAuthHeader)
		}
	case "gotify":
		if data.Gotify != nil {
			var srcUrl, srcToken types.String
			if source != nil && source.Gotify != nil {
				srcUrl = source.Gotify.Url
				srcToken = source.Gotify.Token
			}
			normalizeStringField(&data.Gotify.Url, srcUrl)
			normalizeStringField(&data.Gotify.Token, srcToken)
		}
	case "webpush":
	}
}

func (r *NotificationClientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	disablePayload := `{"enabled":false,"types":0,"options":{}}`
	res, err := r.client.Request(ctx, "POST", notificationPath(r.agent), disablePayload, nil)
	if err != nil {
		if r.notificationDeleteConverged(ctx) {
			return
		}
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		if !strings.Contains(string(res.Body), "Unknown notification agent") && res.StatusCode != 404 {
			resp.Diagnostics.AddWarning("Delete Error", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		}
	}
}

func (r *NotificationClientResource) notificationAgentMissing(ctx context.Context) bool {
	res, err := r.client.Request(ctx, "GET", notificationPath(r.agent), "", nil)
	if err != nil {
		return false
	}
	return res.StatusCode == 404 || strings.Contains(string(res.Body), "Unknown notification agent")
}

func (r *NotificationClientResource) notificationDeleteConverged(ctx context.Context) bool {
	res, err := r.client.Request(ctx, "GET", notificationPath(r.agent), "", nil)
	if err != nil {
		return false
	}
	if res.StatusCode == 404 || strings.Contains(string(res.Body), "Unknown notification agent") {
		return true
	}
	if !StatusIsOK(res.StatusCode) {
		return false
	}

	var payload notificationAgentPayload
	if err := json.Unmarshal(res.Body, &payload); err != nil {
		return false
	}

	return !payload.Enabled && payload.Types == 0
}

func (r *NotificationClientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if req.ID != r.agent {
		resp.Diagnostics.AddError("Invalid import id", fmt.Sprintf("use import id %q for this resource type", r.agent))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
