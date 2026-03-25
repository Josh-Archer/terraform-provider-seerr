// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type ResourceImport struct {
	Name      string
	ExampleID string
	Hint      string
}

var resources = []ResourceImport{
	{Name: "seerr_api_key", ExampleID: "1"},
	{Name: "seerr_api_object", ExampleID: "GET:/api/v1/status"},
	{Name: "seerr_discover_slider", ExampleID: "1"},
	{Name: "seerr_emby_settings", ExampleID: "emby"},
	{Name: "seerr_jellyfin_settings", ExampleID: "jellyfin"},
	{Name: "seerr_job_schedule", ExampleID: "plex-sync", Hint: "The ID of the job, for example: `plex-sync`, `radarr-sync`, etc."},
	{Name: "seerr_main_settings", ExampleID: "main"},
	{Name: "seerr_notification_discord", ExampleID: "discord"},
	{Name: "seerr_notification_slack", ExampleID: "slack"},
	{Name: "seerr_notification_email", ExampleID: "email"},
	{Name: "seerr_notification_lunasea", ExampleID: "lunasea"},
	{Name: "seerr_notification_telegram", ExampleID: "telegram"},
	{Name: "seerr_notification_pushbullet", ExampleID: "pushbullet"},
	{Name: "seerr_notification_pushover", ExampleID: "pushover"},
	{Name: "seerr_notification_ntfy", ExampleID: "ntfy"},
	{Name: "seerr_notification_webhook", ExampleID: "webhook"},
	{Name: "seerr_notification_gotify", ExampleID: "gotify"},
	{Name: "seerr_notification_webpush", ExampleID: "webpush"},
	{Name: "seerr_plex_settings", ExampleID: "plex"},
	{Name: "seerr_radarr_server", ExampleID: "0", Hint: "The ID of the server. For the first server, the ID is `0`."},
	{Name: "seerr_sonarr_server", ExampleID: "0", Hint: "The ID of the server. For the first server, the ID is `0`."},
	{Name: "seerr_tautulli_settings", ExampleID: "tautulli"},
	{Name: "seerr_user", ExampleID: "1", Hint: "You can also import by username or email, e.g., `jdoe` or `jdoe@example.com`."},
	{Name: "seerr_user_permissions", ExampleID: "1", Hint: "The ID of the user."},
	{Name: "seerr_user_settings_permissions", ExampleID: "1", Hint: "The ID of the user."},
	{Name: "seerr_user_watchlist_settings", ExampleID: "1", Hint: "The ID of the user."},
}

func main() {
	baseDir := filepath.Join("examples", "resources")

	for _, res := range resources {
		targetDir := filepath.Join(baseDir, res.Name)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			fmt.Printf("Error creating dir %s: %v\n", targetDir, err)
			continue
		}

		hint := ""
		if res.Hint != "" {
			hint = fmt.Sprintf("\n# %s", res.Hint)
		}

		content := fmt.Sprintf(`# In Terraform 1.5.0 and later, use an import block to import %s. For example:
#
# import {
#   to = %s.example
#   id = "%s"
# }
%s
# Otherwise, use the terraform import command:
terraform import %s.example %s
`, res.Name, res.Name, res.ExampleID, hint, res.Name, res.ExampleID)

		filePath := filepath.Join(targetDir, "import.sh")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			fmt.Printf("Error writing file %s: %v\n", filePath, err)
		} else {
			fmt.Printf("Generated %s\n", filePath)
		}
	}
}
