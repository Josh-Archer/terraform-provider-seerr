package provider

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func notificationAgentResourceOptionAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"discord": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"bot_username":    schema.StringAttribute{Optional: true},
				"bot_avatar_url":  schema.StringAttribute{Optional: true},
				"webhook_url":     schema.StringAttribute{Required: true},
				"enable_mentions": schema.BoolAttribute{Optional: true},
			},
		},
		"slack": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"webhook_url": schema.StringAttribute{Required: true},
			},
		},
		"email": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"email_from":        schema.StringAttribute{Required: true},
				"smtp_host":         schema.StringAttribute{Required: true},
				"smtp_port":         schema.Int64Attribute{Required: true},
				"secure":            schema.BoolAttribute{Optional: true},
				"ignore_tls":        schema.BoolAttribute{Optional: true},
				"require_tls":       schema.BoolAttribute{Optional: true},
				"auth_user":         schema.StringAttribute{Optional: true},
				"auth_pass":         schema.StringAttribute{Optional: true, Sensitive: true},
				"allow_self_signed": schema.BoolAttribute{Optional: true},
				"sender_name":       schema.StringAttribute{Required: true},
				"pgp_private_key":   schema.StringAttribute{Optional: true, Sensitive: true},
				"pgp_password":      schema.StringAttribute{Optional: true, Sensitive: true},
			},
		},
		"lunasea": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"webhook_url":  schema.StringAttribute{Required: true},
				"profile_name": schema.StringAttribute{Optional: true},
			},
		},
		"telegram": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"bot_username":  schema.StringAttribute{Optional: true},
				"bot_api":       schema.StringAttribute{Required: true, Sensitive: true},
				"chat_id":       schema.StringAttribute{Required: true},
				"send_silently": schema.BoolAttribute{Optional: true},
			},
		},
		"pushbullet": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"access_token": schema.StringAttribute{Required: true, Sensitive: true},
				"channel_tag":  schema.StringAttribute{Optional: true},
			},
		},
		"pushover": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"access_token": schema.StringAttribute{Required: true, Sensitive: true},
				"user_token":   schema.StringAttribute{Required: true, Sensitive: true},
				"sound":        schema.StringAttribute{Optional: true},
			},
		},
		"ntfy": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"url":                           schema.StringAttribute{Required: true},
				"topic":                         schema.StringAttribute{Required: true},
				"auth_method_username_password": schema.BoolAttribute{Optional: true},
				"username":                      schema.StringAttribute{Optional: true},
				"password":                      schema.StringAttribute{Optional: true, Sensitive: true},
				"auth_method_token":             schema.BoolAttribute{Optional: true},
				"token":                         schema.StringAttribute{Optional: true, Sensitive: true},
				"priority":                      schema.Int64Attribute{Optional: true},
			},
		},
		"webhook": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"webhook_url":  schema.StringAttribute{Required: true},
				"json_payload": schema.StringAttribute{Required: true},
				"auth_header":  schema.StringAttribute{Optional: true, Sensitive: true},
			},
		},
		"gotify": schema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]schema.Attribute{
				"url":   schema.StringAttribute{Required: true},
				"token": schema.StringAttribute{Required: true, Sensitive: true},
			},
		},
		"webpush": schema.SingleNestedAttribute{
			Optional:   true,
			Attributes: map[string]schema.Attribute{},
		},
	}
}

func notificationAgentDataSourceOptionAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"discord": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"bot_username":    dschema.StringAttribute{Computed: true},
				"bot_avatar_url":  dschema.StringAttribute{Computed: true},
				"webhook_url":     dschema.StringAttribute{Computed: true},
				"enable_mentions": dschema.BoolAttribute{Computed: true},
			},
		},
		"slack": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"webhook_url": dschema.StringAttribute{Computed: true},
			},
		},
		"email": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"email_from":        dschema.StringAttribute{Computed: true},
				"smtp_host":         dschema.StringAttribute{Computed: true},
				"smtp_port":         dschema.Int64Attribute{Computed: true},
				"secure":            dschema.BoolAttribute{Computed: true},
				"ignore_tls":        dschema.BoolAttribute{Computed: true},
				"require_tls":       dschema.BoolAttribute{Computed: true},
				"auth_user":         dschema.StringAttribute{Computed: true},
				"auth_pass":         dschema.StringAttribute{Computed: true, Sensitive: true},
				"allow_self_signed": dschema.BoolAttribute{Computed: true},
				"sender_name":       dschema.StringAttribute{Computed: true},
				"pgp_private_key":   dschema.StringAttribute{Computed: true, Sensitive: true},
				"pgp_password":      dschema.StringAttribute{Computed: true, Sensitive: true},
			},
		},
		"lunasea": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"webhook_url":  dschema.StringAttribute{Computed: true},
				"profile_name": dschema.StringAttribute{Computed: true},
			},
		},
		"telegram": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"bot_username":  dschema.StringAttribute{Computed: true},
				"bot_api":       dschema.StringAttribute{Computed: true, Sensitive: true},
				"chat_id":       dschema.StringAttribute{Computed: true},
				"send_silently": dschema.BoolAttribute{Computed: true},
			},
		},
		"pushbullet": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"access_token": dschema.StringAttribute{Computed: true, Sensitive: true},
				"channel_tag":  dschema.StringAttribute{Computed: true},
			},
		},
		"pushover": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"access_token": dschema.StringAttribute{Computed: true, Sensitive: true},
				"user_token":   dschema.StringAttribute{Computed: true, Sensitive: true},
				"sound":        dschema.StringAttribute{Computed: true},
			},
		},
		"ntfy": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"url":                           dschema.StringAttribute{Computed: true},
				"topic":                         dschema.StringAttribute{Computed: true},
				"auth_method_username_password": dschema.BoolAttribute{Computed: true},
				"username":                      dschema.StringAttribute{Computed: true},
				"password":                      dschema.StringAttribute{Computed: true, Sensitive: true},
				"auth_method_token":             dschema.BoolAttribute{Computed: true},
				"token":                         dschema.StringAttribute{Computed: true, Sensitive: true},
				"priority":                      dschema.Int64Attribute{Computed: true},
			},
		},
		"webhook": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"webhook_url":  dschema.StringAttribute{Computed: true},
				"json_payload": dschema.StringAttribute{Computed: true},
				"auth_header":  dschema.StringAttribute{Computed: true, Sensitive: true},
			},
		},
		"gotify": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"url":   dschema.StringAttribute{Computed: true},
				"token": dschema.StringAttribute{Computed: true, Sensitive: true},
			},
		},
		"webpush": dschema.SingleNestedAttribute{
			Computed:   true,
			Attributes: map[string]dschema.Attribute{},
		},
	}
}

func notificationAgentResourceEventAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"notification_types": schema.SetAttribute{
			Optional:            true,
			Computed:            true,
			ElementType:         types.StringType,
			MarkdownDescription: "Set of notification event types. Valid values include `MEDIA_PENDING`, `MEDIA_APPROVED`, `MEDIA_AVAILABLE`, `MEDIA_FAILED`, `MEDIA_DECLINED`, `MEDIA_AUTO_APPROVED`, `ISSUE_CREATED`, `ISSUE_COMMENT`, `ISSUE_RESOLVED`, `ISSUE_REOPENED`, `MEDIA_AUTO_REQUESTED`.",
		},
	}
}

func notificationAgentDataSourceEventAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"notification_types": dschema.SetAttribute{
			Computed:    true,
			ElementType: types.StringType,
		},
	}
}

func notificationAgentResourceOptionAttribute(agent string) (schema.Attribute, bool) {
	attr, ok := notificationAgentResourceOptionAttributes()[agent]
	return attr, ok
}

func notificationAgentDataSourceOptionAttribute(agent string) (dschema.Attribute, bool) {
	attr, ok := notificationAgentDataSourceOptionAttributes()[agent]
	return attr, ok
}
