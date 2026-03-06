package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func mapFromTypesMap(ctx context.Context, input types.Map) (map[string]string, diag.Diagnostics) {
	out := map[string]string{}
	if input.IsNull() || input.IsUnknown() {
		return out, nil
	}

	var headers map[string]string
	diags := input.ElementsAs(ctx, &headers, false)
	if diags.HasError() {
		return out, diags
	}
	return headers, diags
}

func mergeJSON(base map[string]any, overrideJSON string) (map[string]any, error) {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	if overrideJSON == "" {
		return out, nil
	}

	var override map[string]any
	if err := json.Unmarshal([]byte(overrideJSON), &override); err != nil {
		return nil, err
	}
	for k, v := range override {
		out[k] = v
	}
	return out, nil
}

func findByIDInJSONArray(body []byte, id int64) ([]byte, bool, error) {
	var arr []map[string]any
	if err := json.Unmarshal(body, &arr); err != nil {
		return nil, false, err
	}
	for _, item := range arr {
		raw, ok := item["id"]
		if !ok {
			continue
		}
		switch v := raw.(type) {
		case float64:
			if int64(v) == id {
				b, err := json.Marshal(item)
				return b, true, err
			}
		case int64:
			if v == id {
				b, err := json.Marshal(item)
				return b, true, err
			}
		}
	}
	return nil, false, nil
}

func requireInt64ID(raw string) (int64, error) {
	var v int64
	if _, err := fmt.Sscanf(raw, "%d", &v); err != nil {
		return 0, fmt.Errorf("invalid id %q", raw)
	}
	return v, nil
}
