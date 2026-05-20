resource "seerr_request_retry" "failed" {
  request_id = 42
  trigger    = "operator-retry-1"
}
