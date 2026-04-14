resource "seerr_override_rule" "example" {
  users             = "1,2"
  genre             = "18"
  language          = "en"
  keywords          = "heist"
  profile_id        = 4
  root_folder       = "/media/movies"
  tags              = "1,2"
  radarr_service_id = 1
}
