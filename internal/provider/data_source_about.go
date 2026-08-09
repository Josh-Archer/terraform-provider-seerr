package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &AboutDataSource{}

type AboutDataSource struct {
	client *APIClient
}

type AboutDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Version         types.String `tfsdk:"version"`
	TotalRequests   types.Int64  `tfsdk:"total_requests"`
	TotalMediaItems types.Int64  `tfsdk:"total_media_items"`
	TZ              types.String `tfsdk:"tz"`
	AppDataPath     types.String `tfsdk:"app_data_path"`
}

func NewAboutDataSource() datasource.DataSource { return &AboutDataSource{} }

func (d *AboutDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_about"
}

func (d *AboutDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get Seerr server stats and information.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "The current version of Seerr.",
				Computed:            true,
			},
			"total_requests": schema.Int64Attribute{
				MarkdownDescription: "The total number of requests.",
				Computed:            true,
			},
			"total_media_items": schema.Int64Attribute{
				MarkdownDescription: "The total number of media items.",
				Computed:            true,
			},
			"tz": schema.StringAttribute{
				MarkdownDescription: "The timezone of the server.",
				Computed:            true,
			},
			"app_data_path": schema.StringAttribute{
				MarkdownDescription: "The application data path.",
				Computed:            true,
			},
		},
	}
}

func (d *AboutDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AboutDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AboutDataSourceModel

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/about", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var m map[string]any
	if err := json.Unmarshal(res.Body, &m); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	if v, ok := m["version"].(string); ok {
		data.Version = types.StringValue(v)
	}
	if v, ok := m["totalRequests"].(float64); ok {
		data.TotalRequests = types.Int64Value(int64(v))
	}
	if v, ok := m["totalMediaItems"].(float64); ok {
		data.TotalMediaItems = types.Int64Value(int64(v))
	}
	if v, ok := m["tz"].(string); ok {
		data.TZ = types.StringValue(v)
	}
	if v, ok := m["appDataPath"].(string); ok {
		data.AppDataPath = types.StringValue(v)
	}

	data.ID = types.StringValue("seerr_about")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
