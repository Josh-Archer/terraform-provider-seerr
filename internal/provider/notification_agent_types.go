package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type NotificationAgentDiscordModel struct {
	BotUsername    types.String `tfsdk:"bot_username"`
	BotAvatarUrl   types.String `tfsdk:"bot_avatar_url"`
	WebhookUrl     types.String `tfsdk:"webhook_url"`
	EnableMentions types.Bool   `tfsdk:"enable_mentions"`
}

type NotificationAgentSlackModel struct {
	WebhookUrl types.String `tfsdk:"webhook_url"`
}

type NotificationAgentEmailModel struct {
	EmailFrom       types.String `tfsdk:"email_from"`
	SmtpHost        types.String `tfsdk:"smtp_host"`
	SmtpPort        types.Int64  `tfsdk:"smtp_port"`
	Secure          types.Bool   `tfsdk:"secure"`
	IgnoreTls       types.Bool   `tfsdk:"ignore_tls"`
	RequireTls      types.Bool   `tfsdk:"require_tls"`
	AuthUser        types.String `tfsdk:"auth_user"`
	AuthPass        types.String `tfsdk:"auth_pass"`
	AllowSelfSigned types.Bool   `tfsdk:"allow_self_signed"`
	SenderName      types.String `tfsdk:"sender_name"`
	PgpPrivateKey   types.String `tfsdk:"pgp_private_key"`
	PgpPassword     types.String `tfsdk:"pgp_password"`
}

type NotificationAgentLunaSeaModel struct {
	WebhookUrl  types.String `tfsdk:"webhook_url"`
	ProfileName types.String `tfsdk:"profile_name"`
}

type NotificationAgentTelegramModel struct {
	BotUsername  types.String `tfsdk:"bot_username"`
	BotAPI       types.String `tfsdk:"bot_api"`
	ChatId       types.String `tfsdk:"chat_id"`
	SendSilently types.Bool   `tfsdk:"send_silently"`
}

type NotificationAgentPushbulletModel struct {
	AccessToken types.String `tfsdk:"access_token"`
	ChannelTag  types.String `tfsdk:"channel_tag"`
}

type NotificationAgentPushoverModel struct {
	AccessToken types.String `tfsdk:"access_token"`
	UserToken   types.String `tfsdk:"user_token"`
	Sound       types.String `tfsdk:"sound"`
}

type NotificationAgentNtfyModel struct {
	Url                        types.String `tfsdk:"url"`
	Topic                      types.String `tfsdk:"topic"`
	AuthMethodUsernamePassword types.Bool   `tfsdk:"auth_method_username_password"`
	Username                   types.String `tfsdk:"username"`
	Password                   types.String `tfsdk:"password"`
	AuthMethodToken            types.Bool   `tfsdk:"auth_method_token"`
	Token                      types.String `tfsdk:"token"`
	Priority                   types.Int64  `tfsdk:"priority"`
}

type NotificationAgentWebhookModel struct {
	WebhookUrl  types.String `tfsdk:"webhook_url"`
	JsonPayload types.String `tfsdk:"json_payload"`
	AuthHeader  types.String `tfsdk:"auth_header"`
}

type NotificationAgentGotifyModel struct {
	Url   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

type NotificationAgentWebpushModel struct {
}

type NotificationAgentOptionsModel struct {
	Discord    *NotificationAgentDiscordModel    `tfsdk:"discord"`
	Slack      *NotificationAgentSlackModel      `tfsdk:"slack"`
	Email      *NotificationAgentEmailModel      `tfsdk:"email"`
	LunaSea    *NotificationAgentLunaSeaModel    `tfsdk:"lunasea"`
	Telegram   *NotificationAgentTelegramModel   `tfsdk:"telegram"`
	Pushbullet *NotificationAgentPushbulletModel `tfsdk:"pushbullet"`
	Pushover   *NotificationAgentPushoverModel   `tfsdk:"pushover"`
	Ntfy       *NotificationAgentNtfyModel       `tfsdk:"ntfy"`
	Webhook    *NotificationAgentWebhookModel    `tfsdk:"webhook"`
	Gotify     *NotificationAgentGotifyModel     `tfsdk:"gotify"`
	Webpush    *NotificationAgentWebpushModel    `tfsdk:"webpush"`
}
