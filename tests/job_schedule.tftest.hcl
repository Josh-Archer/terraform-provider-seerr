run "job_schedule_lifecycle" {
  command = apply

  variables {
    job_id   = "plex-sync"
    schedule = "0 0 * * *"
  }

  module {
    source = "./modules/job_schedule"
  }

  assert {
    condition     = seerr_job_schedule.test.job_id == var.job_id
    error_message = "job_id did not match expected value"
  }

  assert {
    condition     = seerr_job_schedule.test.schedule == var.schedule
    error_message = "schedule did not match expected value"
  }
}

run "job_schedule_update" {
  command = apply

  variables {
    job_id   = "plex-sync"
    schedule = "0 1 * * *"
  }

  module {
    source = "./modules/job_schedule"
  }

  assert {
    condition     = seerr_job_schedule.test.job_id == var.job_id
    error_message = "job_id did not match expected value"
  }

  assert {
    condition     = seerr_job_schedule.test.schedule == var.schedule
    error_message = "schedule did not match expected value"
  }
}

