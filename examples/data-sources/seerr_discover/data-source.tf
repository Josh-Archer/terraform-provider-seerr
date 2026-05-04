data "seerr_discover" "movies" {
  media_type   = "movie"
  sort_by      = "popularity.desc"
  watch_region = "US"
  page         = 1
}
