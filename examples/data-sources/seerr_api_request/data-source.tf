data "seerr_api_request" "status" {
  path   = "/api/v1/status"
  method = "GET"
}

output "status_body" {
  value = data.seerr_api_request.status.response_body_json
}

