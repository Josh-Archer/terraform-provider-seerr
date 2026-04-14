variables {
  fixture_job_id = ""
}

run "job_schedule_lifecycle" {
  command = apply

  variables {
    fixture_job_id = var.fixture_job_id
    job_id         = "plex-sync"
    schedule       = "0 0 * * *"
  }

  module {
    source = "./modules/job_schedule"
  }

  assert {
    condition     = output.effective_job_id != ""
    error_message = "job schedule test did not select an available job id"
  }

  assert {
    condition     = seerr_job_schedule.test.schedule == var.schedule
    error_message = "schedule did not match expected value"
  }
}

run "job_schedule_update" {
  command = apply

  variables {
    fixture_job_id = var.fixture_job_id
    job_id         = "plex-sync"
    schedule       = "0 1 * * *"
  }

  module {
    source = "./modules/job_schedule"
  }

  assert {
    condition     = output.effective_job_id != ""
    error_message = "job schedule update did not select an available job id"
  }

  assert {
    condition     = seerr_job_schedule.test.schedule == var.schedule
    error_message = "schedule did not match expected value"
  }
}

