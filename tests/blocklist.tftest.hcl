run "blocklist_lifecycle" {
  command = apply

  variables {
    tmdb_id    = 438631
    media_type = "movie"
    title      = "Dune"
    user_id    = 1
  }

  module {
    source = "./modules/blocklist"
  }

  assert {
    condition     = seerr_blocklist.test.tmdb_id == var.tmdb_id
    error_message = "Blocklist tmdb_id did not match expected value"
  }

  assert {
    condition     = seerr_blocklist.test.media_type == var.media_type
    error_message = "Blocklist media_type did not match expected value"
  }

  assert {
    condition     = seerr_blocklist.test.title == var.title
    error_message = "Blocklist title did not match expected value"
  }

  assert {
    condition     = seerr_blocklist.test.user_id == var.user_id
    error_message = "Blocklist user_id did not match expected value"
  }
}

run "blocklist_no_drift" {
  command = plan
}
