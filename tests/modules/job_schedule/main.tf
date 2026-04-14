terraform {
  required_providers {
    seerr = {
      source  = "josh-archer/seerr"
      version = "99.99.99"
    }
  }
}
variable "job_id" {
  type = string
}

variable "schedule" {
  type = string
}

variable "fixture_job_id" {
  type    = string
  default = ""
}

locals {
  selected_job_id = var.fixture_job_id != "" ? var.fixture_job_id : var.job_id
}

resource "seerr_job_schedule" "test" {
  job_id   = local.selected_job_id
  schedule = var.schedule
}

output "effective_job_id" {
  value = seerr_job_schedule.test.job_id
}
