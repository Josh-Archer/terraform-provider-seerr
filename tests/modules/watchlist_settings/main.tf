terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
variable "user_id" { type = number }
variable "watchlist_sync_movies" { type = bool }
variable "watchlist_sync_tv" { type = bool }

resource "seerr_user_watchlist_settings" "test" {
  user_id               = var.user_id
  watchlist_sync_movies = var.watchlist_sync_movies
  watchlist_sync_tv     = var.watchlist_sync_tv
}

