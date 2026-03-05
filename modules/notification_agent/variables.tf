variable "agent" {
  description = "Notification agent name (discord, slack, telegram, pushbullet, pushover, email, webpush, webhook, gotify, ntfy)."
  type        = string
}

variable "payload" {
  description = "Full payload object for POST /api/v1/settings/notifications/{agent}."
  type        = map(any)
}

variable "skip_delete" {
  description = "If true, do nothing on destroy. If false, send delete_body_json to disable the agent."
  type        = bool
  default     = true
}

variable "delete_body_json" {
  description = "POST payload to send on destroy when skip_delete is false."
  type        = string
  default     = "{\"enabled\":false}"
}
