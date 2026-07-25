---
page_title: "seerr_user_quota Resource - seerr"
subcategory: ""
description: |-
  Manage per-user movie and TV request quotas in Seerr.
---

# seerr_user_quota (Resource)

Manage per-user movie and TV request quotas in Seerr via `/api/v1/user/{userId}/settings/main`.

Setting a quota to `0` means **unlimited** — the user will inherit the global instance default (visible via the `global_*` computed attributes). This resource performs a read-merge-write on the main-settings endpoint so that only the four quota fields are affected; all other per-user settings (locale, region, watchlist sync, etc.) are preserved.

Use [`seerr_permission_set`](../data-sources/permission_set.md) together with `seerr_user_quota` to declaratively manage a user's full access tier (e.g. *standard_user* vs *power_user*).

## Example Usage

### Standard User Quotas

```terraform
# Resolve the permission bitmask for a standard request user
data "seerr_permission_set" "standard" {
  permissions = ["REQUEST", "REQUEST_MOVIE", "REQUEST_TV"]
}

resource "seerr_user" "alice" {
  username    = "alice"
  email       = "alice@example.com"
  permissions = data.seerr_permission_set.standard.permissions
}

# Limit Alice to 5 movies and 3 TV seasons per week
resource "seerr_user_quota" "alice" {
  user_id           = seerr_user.alice.id
  movie_quota_limit = 5
  movie_quota_days  = 7
  tv_quota_limit    = 3
  tv_quota_days     = 7
}
```

### Power User Quotas

```terraform
data "seerr_permission_set" "power" {
  permissions = [
    "REQUEST",
    "REQUEST_MOVIE",
    "REQUEST_TV",
    "AUTO_APPROVE",
    "AUTO_APPROVE_MOVIE",
    "AUTO_APPROVE_TV",
  ]
}

resource "seerr_user" "bob" {
  username    = "bob"
  email       = "bob@example.com"
  permissions = data.seerr_permission_set.power.permissions
}

# Give Bob a generous monthly quota
resource "seerr_user_quota" "bob" {
  user_id           = seerr_user.bob.id
  movie_quota_limit = 20
  movie_quota_days  = 30
  tv_quota_limit    = 10
  tv_quota_days     = 30
}
```

### Unlimited (Inherit Global Default)

```terraform
resource "seerr_user_quota" "unlimited" {
  user_id           = 5
  movie_quota_limit = 0
  movie_quota_days  = 0
  tv_quota_limit    = 0
  tv_quota_days     = 0
}
```

## Argument Reference

- `user_id` (Required, Number) — The numeric ID of the Seerr user whose quotas to manage.
- `movie_quota_limit` (Optional, Computed, Number) — Maximum number of movie requests allowed within `movie_quota_days`. Set to `0` to use the global default.
- `movie_quota_days` (Optional, Computed, Number) — Rolling window in days for the movie quota. Set to `0` to use the global default.
- `tv_quota_limit` (Optional, Computed, Number) — Maximum number of TV season requests allowed within `tv_quota_days`. Set to `0` to use the global default.
- `tv_quota_days` (Optional, Computed, Number) — Rolling window in days for the TV quota. Set to `0` to use the global default.

## Attributes Reference

In addition to the arguments above, the following computed attributes are available:

- `id` — Terraform resource ID (equal to the user_id string).
- `global_movie_quota_limit` — Instance-wide default movie quota limit.
- `global_movie_quota_days` — Instance-wide default movie quota period in days.
- `global_tv_quota_limit` — Instance-wide default TV quota limit.
- `global_tv_quota_days` — Instance-wide default TV quota period in days.

## Import

Import by the numeric user ID:

```shell
terraform import seerr_user_quota.alice 5
```
