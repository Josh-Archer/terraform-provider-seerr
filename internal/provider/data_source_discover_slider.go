package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DiscoverSliderDataSource{}

type DiscoverSliderDataSource struct {
	client *APIClient
}

type DiscoverSliderDataSourceModel struct {
	ID      types.String              `tfsdk:"id"`
	Sliders []DiscoverSliderItemModel `tfsdk:"sliders"`
}

func NewDiscoverSliderDataSource() datasource.DataSource {
	return &DiscoverSliderDataSource{}
}

func (d *DiscoverSliderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_discover_slider"
}

func (d *DiscoverSliderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about discovery sliders in Seerr.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"sliders": schema.ListNestedAttribute{
				MarkdownDescription: "The list of discovery sliders.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The internal ID of the slider.",
						},
						"type": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "The type of the slider (e.g., 1 for TV, 2 for Movie).",
						},
						"is_built_in": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether this is a built-in slider.",
						},
						"enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Whether the slider is enabled.",
						},
						"title": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The title of the slider (only for custom sliders).",
						},
						"data": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The data/query for the slider (only for custom sliders).",
						},
					},
				},
			},
		},
	}
}

func (d *DiscoverSliderDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DiscoverSliderDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DiscoverSliderDataSourceModel

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/discover", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var apiSliders []map[string]any
	if err := json.Unmarshal(res.Body, &apiSliders); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}

	data.Sliders = make([]DiscoverSliderItemModel, 0, len(apiSliders))
	for _, s := range apiSliders {
		item := DiscoverSliderItemModel{
			ID:      d.toInt64(s["id"]),
			Type:    d.toInt64(s["type"]),
			Enabled: d.toBool(s["enabled"]),
		}
		if v, ok := s["isBuiltIn"].(bool); ok {
			item.IsBuiltIn = types.BoolValue(v)
		}
		if v, ok := s["title"].(string); ok {
			item.Title = types.StringValue(v)
		} else {
			item.Title = types.StringNull()
		}
		if v, ok := s["data"].(string); ok {
			item.Data = types.StringValue(v)
		} else {
			item.Data = types.StringNull()
		}
		data.Sliders = append(data.Sliders, item)
	}

	data.ID = types.StringValue("settings")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *DiscoverSliderDataSource) toInt64(v any) types.Int64 {
	switch val := v.(type) {
	case float64:
		return types.Int64Value(int64(val))
	case int64:
		return types.Int64Value(val)
	case int:
		return types.Int64Value(int64(val))
	}
	return types.Int64Null()
}

func (d *DiscoverSliderDataSource) toBool(v any) types.Bool {
	if b, ok := v.(bool); ok {
		return types.BoolValue(b)
	}
	return types.BoolNull()
}
