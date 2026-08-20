run "request_lifecycle" {
  command = apply

  variables {
    media_type = "movie"
    media_id   = 438631
    is_4k      = false
    status     = 1
  }

  module {
    source = "./modules/request"
  }

  assert {
    condition     = seerr_request.test.media_type == var.media_type
    error_message = "Request media_type did not match expected value"
  }

  assert {
    condition     = seerr_request.test.media_id == var.media_id
    error_message = "Request media_id did not match expected value"
  }

  assert {
    condition     = seerr_request.test.is_4k == var.is_4k
    error_message = "Request is_4k did not match expected value"
  }

  assert {
    condition     = seerr_request.test.status == var.status
    error_message = "Request status did not match expected value"
  }
}

run "request_no_drift" {
  command = plan
}
