variable "app_title" {
  description = "Application title."
  type        = string
  default     = null
}

variable "application_url" {
  description = "Public application URL."
  type        = string
  default     = null
}

variable "locale" {
  description = "Application locale."
  type        = string
  default     = null
}

variable "region" {
  description = "Discovery region."
  type        = string
  default     = null
}

variable "original_language" {
  description = "Original language filter."
  type        = string
  default     = null
}

variable "movie_requests_enabled" {
  description = "Whether movie requests are enabled."
  type        = bool
  default     = null
}

variable "series_requests_enabled" {
  description = "Whether series requests are enabled."
  type        = bool
  default     = null
}
