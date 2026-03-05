variable "agent" {
  description = "Notification agent name (discord, slack, telegram, pushbullet, pushover, email, webpush, webhook, gotify, ntfy)."
  type        = string
}

variable "payload" {
  description = "Full payload object for POST /api/v1/settings/notifications/{agent}."
  type        = map(any)
}

variable "skip_delete" {
  description = "Skip DELETE calls for notification settings endpoints."
  type        = bool
  default     = true
}

variable "delete_body_json" {
  description = "Optional POST payload to send on destroy if skip_delete is false."
  type        = string
  default     = "{\"enabled\":false}"
}

