// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestTrimTrailingSlashModifier(t *testing.T) {
	ctx := context.Background()
	mod := TrimTrailingSlashModifier()

	tests := []struct {
		name     string
		planVal  types.String
		expected types.String
	}{
		{
			name:     "Trims trailing slash from URL",
			planVal:  types.StringValue("http://localhost:5055/"),
			expected: types.StringValue("http://localhost:5055"),
		},
		{
			name:     "Trims multiple trailing slashes from URL",
			planVal:  types.StringValue("https://seerr.domain.com///"),
			expected: types.StringValue("https://seerr.domain.com"),
		},
		{
			name:     "Preserves clean URL without trailing slash",
			planVal:  types.StringValue("http://radarr:7878"),
			expected: types.StringValue("http://radarr:7878"),
		},
		{
			name:     "Ignores non-URL string",
			planVal:  types.StringValue("/local/path/"),
			expected: types.StringValue("/local/path/"),
		},
		{
			name:     "Handles null string",
			planVal:  types.StringNull(),
			expected: types.StringNull(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				PlanValue: tt.planVal,
			}
			resp := &planmodifier.StringResponse{
				PlanValue: tt.planVal,
			}
			mod.PlanModifyString(ctx, req, resp)
			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}

func TestJSONCanonicalModifier(t *testing.T) {
	ctx := context.Background()
	mod := JSONCanonicalModifier()

	tests := []struct {
		name     string
		stateVal types.String
		planVal  types.String
		expected types.String
	}{
		{
			name:     "Suppresses formatting diff for identical JSON objects",
			stateVal: types.StringValue(`{"a":1,"b":2}`),
			planVal:  types.StringValue("{\n  \"b\": 2,\n  \"a\": 1\n}"),
			expected: types.StringValue(`{"a":1,"b":2}`),
		},
		{
			name:     "Maintains plan value when JSON is actually different",
			stateVal: types.StringValue(`{"a":1}`),
			planVal:  types.StringValue(`{"a":2}`),
			expected: types.StringValue(`{"a":2}`),
		},
		{
			name:     "Handles invalid JSON without error",
			stateVal: types.StringValue("invalid json"),
			planVal:  types.StringValue("different invalid json"),
			expected: types.StringValue("different invalid json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				StateValue: tt.stateVal,
				PlanValue:  tt.planVal,
			}
			resp := &planmodifier.StringResponse{
				PlanValue: tt.planVal,
			}
			mod.PlanModifyString(ctx, req, resp)
			assert.Equal(t, tt.expected, resp.PlanValue)
		})
	}
}
