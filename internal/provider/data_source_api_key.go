package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &APIKeyDataSource{}

type APIKeyDataSource struct {
	client *APIClient
}

type APIKeyDataSourceModel struct {
	ApiKey     types.String `tfsdk:"api_key"`
	StatusCode types.Int64  `tfsdk:"status_code"`
}

func NewAPIKeyDataSource() datasource.DataSource { return &APIKeyDataSource{} }

func (d *APIKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_key"
}

func (d *APIKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read the Seerr API key via /api/v1/settings/main.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The current Seerr API key.",
				Computed:            true,
				Sensitive:           true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code.",
				Computed:            true,
			},
		},
	}
}

func (d *APIKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *APIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APIKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/main", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}

	apiKey, ok := decoded["apiKey"].(string)
	if !ok {
		resp.Diagnostics.AddError("Read Failed", "apiKey not found in response")
		return
	}

	data.ApiKey = types.StringValue(apiKey)
	data.StatusCode = types.Int64Value(int64(res.StatusCode))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
