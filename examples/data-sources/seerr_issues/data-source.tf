data "seerr_issues" "all" {}

output "total_issues" {
  value = length(data.seerr_issues.all.issues)
}
