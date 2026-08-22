// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TrimTrailingSlashModifier returns a StringPlanModifier that strips trailing slashes
// from HTTP/HTTPS URLs to eliminate semantic diffs between configured and returned values.
func TrimTrailingSlashModifier() planmodifier.String {
	return trimTrailingSlashModifier{}
}

type trimTrailingSlashModifier struct{}

func (m trimTrailingSlashModifier) Description(_ context.Context) string {
	return "Normalizes URLs by trimming trailing slashes to prevent unnecessary diffs"
}

func (m trimTrailingSlashModifier) MarkdownDescription(_ context.Context) string {
	return "Normalizes URLs by trimming trailing slashes to prevent unnecessary diffs."
}

func (m trimTrailingSlashModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}
	val := req.PlanValue.ValueString()
	if strings.HasSuffix(val, "/") && (strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://")) {
		trimmed := strings.TrimRight(val, "/")
		if trimmed != "" {
			resp.PlanValue = types.StringValue(trimmed)
		}
	}
}

// JSONCanonicalModifier returns a StringPlanModifier that suppresses diffs
// between semantically identical JSON documents that differ only in formatting or key ordering.
func JSONCanonicalModifier() planmodifier.String {
	return jsonCanonicalModifier{}
}

type jsonCanonicalModifier struct{}

func (m jsonCanonicalModifier) Description(_ context.Context) string {
	return "Suppresses formatting and whitespace diffs in JSON strings when semantically equivalent"
}

func (m jsonCanonicalModifier) MarkdownDescription(_ context.Context) string {
	return "Suppresses formatting and whitespace diffs in JSON strings when semantically equivalent."
}

func (m jsonCanonicalModifier) PlanModifyString(_ context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	planJSON := strings.TrimSpace(req.PlanValue.ValueString())
	stateJSON := strings.TrimSpace(req.StateValue.ValueString())
	if planJSON == "" || stateJSON == "" || planJSON == stateJSON {
		return
	}

	var planObj, stateObj any
	if err1 := json.Unmarshal([]byte(planJSON), &planObj); err1 == nil {
		if err2 := json.Unmarshal([]byte(stateJSON), &stateObj); err2 == nil {
			if reflect.DeepEqual(planObj, stateObj) {
				resp.PlanValue = req.StateValue
			}
		}
	}
}
