package provider

import (
	"context"

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
