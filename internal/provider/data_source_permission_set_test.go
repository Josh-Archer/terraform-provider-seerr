package provider

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccPermissionSetDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
data "seerr_permission_set" "test" {
  request       = true
  request_4k    = true
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.seerr_permission_set.test", "permissions", "1056"),
				),
			},
		},
	})
}

func TestPermissionSet_Consistency(t *testing.T) {
	content, err := os.ReadFile("../../seerr_permissions.ts")
	if err != nil {
		t.Fatalf("failed to read seerr_permissions.ts: %v", err)
	}

	contentStr := string(content)

	// Regex to extract enum values.
	enumRegex := regexp.MustCompile(`(?m)^\s*([A-Z0-9_]+)\s*=\s*([0-9]+)\s*,?`)

	matches := enumRegex.FindAllStringSubmatch(contentStr, -1)

	parsedPermissions := make(map[string]int64)
	for _, match := range matches {
		if len(match) == 3 {
			key := strings.ToLower(match[1])
			if key == "none" {
				continue
			}
			val, err := strconv.ParseInt(match[2], 10, 64)
			if err != nil {
				t.Fatalf("failed to parse int for key %s: %v", key, err)
			}
			parsedPermissions[key] = val
		}
	}

	assert.Equal(t, len(parsedPermissions), len(PermissionsMap), "Number of permissions in TS file does not match provider map")

	for k, v := range parsedPermissions {
		providerVal, ok := PermissionsMap[k]
		assert.True(t, ok, "Permission %s found in TS file but missing in provider map", k)
		assert.Equal(t, v, providerVal, "Permission value mismatch for %s", k)
	}

	for k, v := range PermissionsMap {
		tsVal, ok := parsedPermissions[k]
		assert.True(t, ok, "Permission %s found in provider map but missing in TS file", k)
		assert.Equal(t, v, tsVal, "Permission value mismatch for %s", k)
	}
}
