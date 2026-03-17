// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var resources = map[string]string{
	"seerr_api_key":                   "1",
	"seerr_api_object":                "GET:/api/v1/status",
	"seerr_discover_slider":           "1",
	"seerr_emby_settings":             "emby",
	"seerr_jellyfin_settings":         "jellyfin",
	"seerr_main_settings":             "main",
	"seerr_notification_agent":        "discord",
	"seerr_plex_settings":             "plex",
	"seerr_radarr_server":             "0",
	"seerr_sonarr_server":             "0",
	"seerr_tautulli_settings":         "tautulli",
	"seerr_user":                      "1\n\n# Additionally, you can import by username or email:\n# terraform import seerr_user.example jdoe\n# terraform import seerr_user.example jdoe@example.com",
	"seerr_user_permissions":          "1",
	"seerr_user_settings_permissions": "1",
	"seerr_user_watchlist_settings":   "1",
}

func main() {
	baseDir := filepath.Join("examples", "resources")

	for resName, idExample := range resources {
		targetDir := filepath.Join(baseDir, resName)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			fmt.Printf("Error creating dir %s: %v\n", targetDir, err)
			continue
		}

		content := fmt.Sprintf(`# In Terraform 1.5.0 and later, use an import block to import %s. For example:
#
# import {
#   to = %s.example
#   id = "%s"
# }
#
# Otherwise, use the terraform import command:
terraform import %s.example %s
`, resName, resName, idExample, resName, idExample)

		filePath := filepath.Join(targetDir, "import.sh")
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			fmt.Printf("Error writing file %s: %v\n", filePath, err)
		} else {
			fmt.Printf("Generated %s\n", filePath)
		}
	}
}
