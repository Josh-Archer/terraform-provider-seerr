resource "seerr_sonarr_server" "example" {
  name               = "Sonarr"
  hostname           = "sonarr.local"
  port               = 8989
  api_key            = "your-sonarr-api-key"
  quality_profile_id = 1
  active_directory   = "/tv"
  is_default         = true
}
