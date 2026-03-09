resource "seerr_main_settings" "main" {
  app_title              = "Seerr"
  application_url        = "https://seerr.example.com"
  locale                 = "en"
  movie_requests_enabled = true
  series_requests_enabled = true
}
