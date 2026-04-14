terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}

variable "notification_agents_supported" {
  type    = bool
  default = true
}

data "seerr_notification_agents" "test" {
  count = var.notification_agents_supported ? 1 : 0
}

output "supported" {
  value = var.notification_agents_supported
}

output "agent_count" {
  value = var.notification_agents_supported ? length(data.seerr_notification_agents.test[0].agents) : 0
}
