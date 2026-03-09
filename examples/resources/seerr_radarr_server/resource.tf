resource "seerr_radarr_server" "example" {
  name               = "Radarr"
  hostname           = "radarr.local"
  port               = 7878
  api_key            = "your-radarr-api-key"
  quality_profile_id = 1
  active_directory   = "/movies"
  is_default         = true
}
