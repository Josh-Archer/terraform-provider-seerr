package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

func useStateForUnknownBool() []planmodifier.Bool {
	return []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}
}

func optionalComputedBoolAttr() schema.BoolAttribute {
	return schema.BoolAttribute{
		Optional:      true,
		Computed:      true,
		PlanModifiers: useStateForUnknownBool(),
	}
}
