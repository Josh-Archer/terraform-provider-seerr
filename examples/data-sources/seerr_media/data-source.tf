data "seerr_media" "available" {
  filter = "available"
  sort   = "mediaAdded"
  take   = 50
}
