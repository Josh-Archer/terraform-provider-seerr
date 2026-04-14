resource "seerr_tautulli_settings" "example" {
  hostname     = "tautulli.internal"
  port         = 8181
  api_key      = var.tautulli_api_key
  use_ssl      = false
  external_url = "https://tautulli.example.com"
}
