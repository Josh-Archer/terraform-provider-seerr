resource "seerr_main_settings" "main" {
  app_title               = "Seerr"
  application_url         = "https://seerr.example.com"
  locale                  = "en"
  default_permissions     = 32
  version_check           = true
  movie_requests_enabled  = true
  series_requests_enabled = true
}
