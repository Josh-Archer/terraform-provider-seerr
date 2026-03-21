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

resource "seerr_job_schedule" "test" {
  job_id   = var.job_id
  schedule = var.schedule
}
