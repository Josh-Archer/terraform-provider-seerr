# Upstream Seerr Reference Snapshots

These files match [Seerr commit ff4ab21](https://github.com/seerr-team/seerr/tree/ff4ab21a9e664082feb954c6f482b512475f1fa6)
after normalizing text encoding, line endings, and trailing whitespace.

| Snapshot | Upstream path |
| --- | --- |
| `UserSettings.ts` | `server/entity/UserSettings.ts` |
| `userInterfaces.ts` | `server/interfaces/api/userInterfaces.ts` |
| `seerr_permissions.ts` | `server/lib/permissions.ts` |
| `seerr_userSettingsInterfaces.ts` | `server/interfaces/api/userSettingsInterfaces.ts` |
| `seerr_usersettings_routes.ts` | `server/routes/user/usersettings.ts` |

`tools/gen_permissions.go` reads `seerr_permissions.ts` when running
`go generate ./...` to update `internal/provider/permissions_gen.go`.
