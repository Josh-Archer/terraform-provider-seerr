package provider

import (
	"context"
	"fmt"

	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &BlocklistDataSource{}

type BlocklistDataSource struct {
	client *APIClient
}

func NewBlocklistDataSource() datasource.DataSource { return &BlocklistDataSource{} }

func (d *BlocklistDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_blocklist"
}

func (d *BlocklistDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := blocklistDataSourceAttributes()
	attrs["media_type"] = schema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.OneOf("movie", "tv"),
		},
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read a Seerr blocklist entry by TMDB ID and media type.",
		Attributes:          attrs,
	}
}

func (d *BlocklistDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BlocklistDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BlocklistModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resource := &BlocklistResource{client: d.client}
	if err := resource.refreshBlocklist(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	if data.UserID.IsNull() {
		data.UserID = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
