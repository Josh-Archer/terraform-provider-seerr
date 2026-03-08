terraform {
  required_providers {
    seerr = {
      source = "josh-archer/seerr"
    }
  }
}

variable "user_id" { type = number }
variable "global_watchlist_sync" { type = bool }

resource "seerr_user_watchlist_settings" "test" {
  user_id               = var.user_id
  global_watchlist_sync = var.global_watchlist_sync
}

