data "seerr_requests" "all" {}

output "total_requests" {
  value = length(data.seerr_requests.all.requests)
}
