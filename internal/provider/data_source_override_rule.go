package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

var _ datasource.DataSource = &OverrideRuleDataSource{}

type OverrideRuleDataSource struct {
	client *APIClient
}

func NewOverrideRuleDataSource() datasource.DataSource { return &OverrideRuleDataSource{} }

func (d *OverrideRuleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_override_rule"
}

func (d *OverrideRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read a Seerr override rule by ID via `/api/v1/overrideRule`.",
		Attributes:          overrideRuleDataSourceSchema(),
	}
}

func (d *OverrideRuleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *OverrideRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OverrideRuleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource := &OverrideRuleResource{client: d.client}
	found, err := resource.fetchOverrideRule(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if found == nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("override rule %q not found", data.ID.ValueString()))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, found)...)
}
