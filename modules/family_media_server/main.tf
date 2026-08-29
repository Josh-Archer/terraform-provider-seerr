terraform {
  required_version = ">= 1.5.0"

  required_providers {
    seerr = {
      source = "registry.opentofu.org/josh-archer/seerr"
    }
  }
}

data "seerr_permission_set" "standard" {
  request       = true
  request_movie = true
  request_tv    = true
}

data "seerr_permission_set" "power" {
  request       = true
  request_movie = true
  request_tv    = true
  request_4k    = true
  auto_approve  = true
}

data "seerr_permission_set" "admin" {
  admin = true
}

locals {
  permission_sets = {
    standard = data.seerr_permission_set.standard.permissions
    power    = data.seerr_permission_set.power.permissions
    admin    = data.seerr_permission_set.admin.permissions
  }
}

resource "seerr_user" "this" {
  for_each = var.users

  username          = each.value.username
  email             = each.value.email
  permissions       = coalesce(each.value.permissions, local.permission_sets[each.value.permission_set])
  locale            = each.value.locale
  discover_region   = each.value.discover_region
  streaming_region  = each.value.streaming_region
  original_language = each.value.original_language

  dynamic "notification_settings" {
    for_each = each.value.notification_settings == null ? [] : [each.value.notification_settings]
    content {
      email_enabled              = notification_settings.value.email_enabled
      discord_enabled            = notification_settings.value.discord_enabled
      discord_id                 = notification_settings.value.discord_id
      telegram_enabled           = notification_settings.value.telegram_enabled
      telegram_chat_id           = notification_settings.value.telegram_chat_id
      telegram_bot_username      = notification_settings.value.telegram_bot_username
      telegram_message_thread_id = notification_settings.value.telegram_message_thread_id
      telegram_send_silently     = notification_settings.value.telegram_send_silently
      pushbullet_access_token    = notification_settings.value.pushbullet_access_token
      pushover_application_token = notification_settings.value.pushover_application_token
      pushover_user_key          = notification_settings.value.pushover_user_key
      pushover_sound             = notification_settings.value.pushover_sound
      pgp_key                    = notification_settings.value.pgp_key

      dynamic "notification_types" {
        for_each = notification_settings.value.notification_types == null ? [] : [notification_settings.value.notification_types]
        content {
          email      = notification_types.value.email
          discord    = notification_types.value.discord
          telegram   = notification_types.value.telegram
          pushbullet = notification_types.value.pushbullet
          pushover   = notification_types.value.pushover
          ntfy       = notification_types.value.ntfy
          gotify     = notification_types.value.gotify
          slack      = notification_types.value.slack
          webhook    = notification_types.value.webhook
          webpush    = notification_types.value.webpush
        }
      }
    }
  }
}

resource "seerr_user_quota" "this" {
  for_each = {
    for key, user in var.users : key => user
    if user.quota != null
  }

  user_id           = tonumber(seerr_user.this[each.key].id)
  movie_quota_limit = each.value.quota.movie_limit
  movie_quota_days  = each.value.quota.movie_days
  tv_quota_limit    = each.value.quota.tv_limit
  tv_quota_days     = each.value.quota.tv_days
}

resource "seerr_discover_slider" "this" {
  count = length(var.discover_sliders) == 0 ? 0 : 1

  dynamic "sliders" {
    for_each = var.discover_sliders
    content {
      type    = sliders.value.type
      enabled = sliders.value.enabled
      title   = sliders.value.title
      data    = sliders.value.data
    }
  }
}

