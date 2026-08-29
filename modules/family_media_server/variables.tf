variable "users" {
  description = "Family users keyed by a stable logical name. Permission sets are standard, power, or admin; an explicit numeric bitmask takes precedence."
  type = map(object({
    username          = string
    email             = string
    permission_set    = optional(string, "standard")
    permissions       = optional(number)
    locale            = optional(string)
    discover_region   = optional(string)
    streaming_region  = optional(string)
    original_language = optional(string)
    quota = optional(object({
      movie_limit = optional(number, 0)
      movie_days  = optional(number, 0)
      tv_limit    = optional(number, 0)
      tv_days     = optional(number, 0)
    }))
    notification_settings = optional(object({
      email_enabled              = optional(bool)
      discord_enabled            = optional(bool)
      discord_id                 = optional(string)
      telegram_enabled           = optional(bool)
      telegram_chat_id           = optional(string)
      telegram_bot_username      = optional(string)
      telegram_message_thread_id = optional(string)
      telegram_send_silently     = optional(bool)
      pushbullet_access_token    = optional(string)
      pushover_application_token = optional(string)
      pushover_user_key          = optional(string)
      pushover_sound             = optional(string)
      pgp_key                    = optional(string)
      notification_types = optional(object({
        email      = optional(number)
        discord    = optional(number)
        telegram   = optional(number)
        pushbullet = optional(number)
        pushover   = optional(number)
        ntfy       = optional(number)
        gotify     = optional(number)
        slack      = optional(number)
        webhook    = optional(number)
        webpush    = optional(number)
      }))
    }))
  }))

  validation {
    condition = alltrue([
      for user in values(var.users) : contains(["standard", "power", "admin"], user.permission_set)
    ])
    error_message = "Each permission_set must be standard, power, or admin."
  }
}

variable "discover_sliders" {
  description = "Ordered discover sliders. Leave empty to avoid managing the singleton slider configuration."
  type = list(object({
    type    = number
    enabled = optional(bool, true)
    title   = optional(string)
    data    = optional(string)
  }))
  default = []
}

