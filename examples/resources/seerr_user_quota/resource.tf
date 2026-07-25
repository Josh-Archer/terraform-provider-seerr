resource "seerr_user_quota" "standard_user" {
  user_id           = 5
  movie_quota_limit = 5
  movie_quota_days  = 7
  tv_quota_limit    = 3
  tv_quota_days     = 7
}
