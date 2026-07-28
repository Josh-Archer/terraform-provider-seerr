package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNormalizeNotificationModelAllAgentsOmittedAndPartialFields(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("Discord Omitted Fields", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("discord")}
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"webhookUrl":"https://discord.com/api/webhooks/123","botUsername":"","botAvatarUrl":null}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		// Source has null for optional fields (omitted in HCL)
		source := &NotificationAgentModel{
			Agent:   types.StringValue("discord"),
			Discord: &NotificationAgentDiscordModel{WebhookUrl: types.StringValue("https://discord.com/api/webhooks/123")},
		}
		normalizeNotificationModel(data, source)

		if !data.Discord.BotUsername.IsNull() {
			t.Errorf("expected BotUsername to be null, got %#v", data.Discord.BotUsername)
		}
		if !data.Discord.BotAvatarUrl.IsNull() {
			t.Errorf("expected BotAvatarUrl to be null, got %#v", data.Discord.BotAvatarUrl)
		}
		if data.Discord.EnableMentions.IsNull() || data.Discord.EnableMentions.ValueBool() {
			t.Errorf("expected EnableMentions to be false, got %#v", data.Discord.EnableMentions)
		}
		if !data.NotificationTypes.IsNull() {
			t.Errorf("expected NotificationTypes to be null when omitted, got %#v", data.NotificationTypes)
		}
	})

	t.Run("Email Sensitive and Optional Fields Preservation", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("email")}
		// API returns empty strings for sensitive fields and omitted fields
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"emailFrom":"no-reply@example.com","smtpHost":"smtp.example.com","smtpPort":587,"senderName":"Seerr","authPass":"","pgpPrivateKey":"","pgpPassword":""}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		source := &NotificationAgentModel{
			Agent: types.StringValue("email"),
			Email: &NotificationAgentEmailModel{
				EmailFrom:     types.StringValue("no-reply@example.com"),
				SmtpHost:      types.StringValue("smtp.example.com"),
				SmtpPort:      types.Int64Value(587),
				SenderName:    types.StringValue("Seerr"),
				AuthPass:      types.StringValue("my-secret-smtp-password"),
				PgpPrivateKey: types.StringValue("-----BEGIN PGP PRIVATE KEY BLOCK-----"),
				PgpPassword:   types.StringValue("pgp-secret"),
			},
		}
		normalizeNotificationModel(data, source)

		if got := data.Email.AuthPass.ValueString(); got != "my-secret-smtp-password" {
			t.Errorf("expected AuthPass to be preserved, got %q", got)
		}
		if got := data.Email.PgpPrivateKey.ValueString(); got != "-----BEGIN PGP PRIVATE KEY BLOCK-----" {
			t.Errorf("expected PgpPrivateKey to be preserved, got %q", got)
		}
		if got := data.Email.PgpPassword.ValueString(); got != "pgp-secret" {
			t.Errorf("expected PgpPassword to be preserved, got %q", got)
		}
		if !data.Email.AuthUser.IsNull() {
			t.Errorf("expected AuthUser to be null when omitted, got %#v", data.Email.AuthUser)
		}
		if data.Email.Secure.IsNull() || data.Email.Secure.ValueBool() {
			t.Errorf("expected Secure to be false, got %#v", data.Email.Secure)
		}
	})

	t.Run("Telegram Sensitive API Key and Optional Fields", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("telegram")}
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"chatId":"123456","botUsername":"","botAPI":""}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		source := &NotificationAgentModel{
			Agent: types.StringValue("telegram"),
			Telegram: &NotificationAgentTelegramModel{
				ChatId: types.StringValue("123456"),
				BotAPI: types.StringValue("123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"),
			},
		}
		normalizeNotificationModel(data, source)

		if got := data.Telegram.BotAPI.ValueString(); got != "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11" {
			t.Errorf("expected BotAPI to be preserved, got %q", got)
		}
		if !data.Telegram.BotUsername.IsNull() {
			t.Errorf("expected BotUsername to be null, got %#v", data.Telegram.BotUsername)
		}
		if data.Telegram.SendSilently.IsNull() || data.Telegram.SendSilently.ValueBool() {
			t.Errorf("expected SendSilently to be false, got %#v", data.Telegram.SendSilently)
		}
	})

	t.Run("Ntfy Priority and Token Omission", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("ntfy")}
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"url":"https://ntfy.sh","topic":"my-topic","priority":0,"token":""}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		source := &NotificationAgentModel{
			Agent: types.StringValue("ntfy"),
			Ntfy: &NotificationAgentNtfyModel{
				Url:   types.StringValue("https://ntfy.sh"),
				Topic: types.StringValue("my-topic"),
				Token: types.StringValue("tk_secret123"),
			},
		}
		normalizeNotificationModel(data, source)

		if got := data.Ntfy.Token.ValueString(); got != "tk_secret123" {
			t.Errorf("expected Token to be preserved, got %q", got)
		}
		if !data.Ntfy.Priority.IsNull() {
			t.Errorf("expected Priority to be null when source priority is null and API returned 0, got %#v", data.Ntfy.Priority)
		}
		if !data.Ntfy.Username.IsNull() {
			t.Errorf("expected Username to be null when omitted, got %#v", data.Ntfy.Username)
		}
	})

	t.Run("Pushover Tokens and Sound Omission", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("pushover")}
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"accessToken":"","userToken":"","sound":""}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		source := &NotificationAgentModel{
			Agent: types.StringValue("pushover"),
			Pushover: &NotificationAgentPushoverModel{
				AccessToken: types.StringValue("acc_token_secret"),
				UserToken:   types.StringValue("usr_token_secret"),
			},
		}
		normalizeNotificationModel(data, source)

		if got := data.Pushover.AccessToken.ValueString(); got != "acc_token_secret" {
			t.Errorf("expected AccessToken to be preserved, got %q", got)
		}
		if got := data.Pushover.UserToken.ValueString(); got != "usr_token_secret" {
			t.Errorf("expected UserToken to be preserved, got %q", got)
		}
		if !data.Pushover.Sound.IsNull() {
			t.Errorf("expected Sound to be null when omitted, got %#v", data.Pushover.Sound)
		}
	})

	t.Run("Pushbullet Access Token and Channel Tag", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("pushbullet")}
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"accessToken":"","channelTag":""}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		source := &NotificationAgentModel{
			Agent: types.StringValue("pushbullet"),
			Pushbullet: &NotificationAgentPushbulletModel{
				AccessToken: types.StringValue("push_token_secret"),
			},
		}
		normalizeNotificationModel(data, source)

		if got := data.Pushbullet.AccessToken.ValueString(); got != "push_token_secret" {
			t.Errorf("expected AccessToken to be preserved, got %q", got)
		}
		if !data.Pushbullet.ChannelTag.IsNull() {
			t.Errorf("expected ChannelTag to be null, got %#v", data.Pushbullet.ChannelTag)
		}
	})

	t.Run("Webhook Auth Header and Payload", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("webhook")}
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"webhookUrl":"https://example.com/wh","jsonPayload":"{}","authHeader":""}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		source := &NotificationAgentModel{
			Agent: types.StringValue("webhook"),
			Webhook: &NotificationAgentWebhookModel{
				WebhookUrl:  types.StringValue("https://example.com/wh"),
				JsonPayload: types.StringValue("{}"),
				AuthHeader:  types.StringValue("Bearer secret_token"),
			},
		}
		normalizeNotificationModel(data, source)

		if got := data.Webhook.AuthHeader.ValueString(); got != "Bearer secret_token" {
			t.Errorf("expected AuthHeader to be preserved, got %q", got)
		}
	})

	t.Run("Gotify Token Preservation", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("gotify")}
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"url":"https://gotify.example.com","token":""}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		source := &NotificationAgentModel{
			Agent: types.StringValue("gotify"),
			Gotify: &NotificationAgentGotifyModel{
				Url:   types.StringValue("https://gotify.example.com"),
				Token: types.StringValue("gotify_secret_token"),
			},
		}
		normalizeNotificationModel(data, source)

		if got := data.Gotify.Token.ValueString(); got != "gotify_secret_token" {
			t.Errorf("expected Token to be preserved, got %q", got)
		}
	})

	t.Run("LunaSea Profile Name Omission", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("lunasea")}
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"webhookUrl":"https://notify.lunasea.app/123","profileName":""}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		source := &NotificationAgentModel{
			Agent:   types.StringValue("lunasea"),
			LunaSea: &NotificationAgentLunaSeaModel{WebhookUrl: types.StringValue("https://notify.lunasea.app/123")},
		}
		normalizeNotificationModel(data, source)

		if !data.LunaSea.ProfileName.IsNull() {
			t.Errorf("expected ProfileName to be null, got %#v", data.LunaSea.ProfileName)
		}
	})

	t.Run("Slack Webhook URL", func(t *testing.T) {
		t.Parallel()
		data := &NotificationAgentModel{Agent: types.StringValue("slack")}
		body := []byte(`{"enabled":true,"embedPoster":false,"types":0,"options":{"webhookUrl":"https://hooks.slack.com/services/123"}}`)
		if err := parsePayload(ctx, data, body); err != nil {
			t.Fatalf("parsePayload failed: %v", err)
		}
		source := &NotificationAgentModel{
			Agent: types.StringValue("slack"),
			Slack: &NotificationAgentSlackModel{WebhookUrl: types.StringValue("https://hooks.slack.com/services/123")},
		}
		normalizeNotificationModel(data, source)

		if got := data.Slack.WebhookUrl.ValueString(); got != "https://hooks.slack.com/services/123" {
			t.Errorf("expected WebhookUrl to be set, got %q", got)
		}
	})
}
