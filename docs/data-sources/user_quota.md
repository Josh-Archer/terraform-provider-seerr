---
page_title: "seerr_user_quota Data Source - seerr"
subcategory: ""
description: |-
  Read per-user movie and TV request quotas from Seerr.
---

# seerr_user_quota (Data Source)

Read per-user movie and TV request quotas from Seerr via `/api/v1/user/{userId}/settings/main`.

A quota value of `0` means the user inherits the global instance default. Use `global_movie_quota_limit` / `global_tv_quota_limit` to inspect those defaults for comparison.

## Example Usage

```terraform
# Read quota settings for user 5
data "seerr_user_quota" "alice" {
  user_id = 5
}

output "alice_movie_quota" {
  value = data.seerr_user_quota.alice.movie_quota_limit
}

output "alice_global_movie_quota" {
  description = "Global movie quota limit for reference"
  value       = data.seerr_user_quota.alice.global_movie_quota_limit
}
```

## Argument Reference

- `user_id` (Required, Number) — The numeric ID of the user to look up.

## Attributes Reference

- `movie_quota_limit` — Per-user movie request quota limit (`0` = inherit global).
- `movie_quota_days` — Per-user movie quota rolling window in days (`0` = inherit global).
- `tv_quota_limit` — Per-user TV request quota limit (`0` = inherit global).
- `tv_quota_days` — Per-user TV quota rolling window in days (`0` = inherit global).
- `global_movie_quota_limit` — Instance-wide default movie quota limit.
- `global_movie_quota_days` — Instance-wide default movie quota period in days.
- `global_tv_quota_limit` — Instance-wide default TV quota limit.
- `global_tv_quota_days` — Instance-wide default TV quota period in days.
