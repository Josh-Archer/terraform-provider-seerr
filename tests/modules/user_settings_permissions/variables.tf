variable "user_id" {
  type    = number
  default = null
}

variable "email" {
  type    = string
  default = null
}

variable "username" {
  type    = string
  default = null
}

variable "auto_approve_movies" {
  type    = bool
  default = false
}

variable "auto_approve_tv" {
  type    = bool
  default = false
}

variable "auto_approve_4k_movies" {
  type    = bool
  default = false
}

variable "auto_approve_4k_tv" {
  type    = bool
  default = false
}
