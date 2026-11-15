# Upstream Seerr Reference Snapshots

These TypeScript files are reference snapshots copied from upstream Seerr/Overseerr (`seerr-team/seerr`) to preserve API permission masks and user/settings type definitions:

- `seerr_permissions.ts` — source input for `go generate ./...` (`tools/gen_permissions.go` -> `internal/provider/permissions_gen.go`).
- `UserSettings.ts`, `userInterfaces.ts`, `seerr_userSettingsInterfaces.ts`, `seerr_user_routes.ts`, `seerr_usersettings_routes.ts` — type/route reference snapshots retained alongside OpenAPI specs.
