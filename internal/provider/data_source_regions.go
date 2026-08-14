package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RegionsDataSource{}

type RegionsDataSource struct {
	client *APIClient
}

type RegionsDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Regions types.List   `tfsdk:"regions"`
	Total   types.Int64  `tfsdk:"total"`
}

type RegionModel struct {
	ISO31661    types.String `tfsdk:"iso_3166_1"`
	EnglishName types.String `tfsdk:"english_name"`
	Name        types.String `tfsdk:"name"`
}

func NewRegionsDataSource() datasource.DataSource {
	return &RegionsDataSource{}
}

func (d *RegionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *RegionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch supported ISO country and region codes from TMDB for regional watch provider filtering.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier.",
				Computed:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of regions returned.",
				Computed:            true,
			},
			"regions": schema.ListNestedAttribute{
				MarkdownDescription: "List of country/region codes.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"iso_3166_1": schema.StringAttribute{
							MarkdownDescription: "ISO 3166-1 country/region code (e.g. `US`, `GB`, `CA`).",
							Computed:            true,
						},
						"english_name": schema.StringAttribute{
							MarkdownDescription: "English name of the country/region.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Native name of the country/region.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *RegionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *APIClient, got: %T", req.ProviderData))
		return
	}
	d.client = client
}

type rawRegionResponse struct {
	ISO31661    string `json:"iso_3166_1"`
	EnglishName string `json:"english_name"`
	Name        string `json:"name"`
}

func (d *RegionsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured API Client", "API client was not provided during configuration.")
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/regions", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var raw []rawRegionResponse
	if err := json.Unmarshal(res.Body, &raw); err != nil {
		resp.Diagnostics.AddError("Error Parsing Regions Response", fmt.Sprintf("Failed to parse response: %s", err))
		return
	}

	regionElements := make([]RegionModel, 0, len(raw))
	for _, r := range raw {
		regionElements = append(regionElements, RegionModel{
			ISO31661:    types.StringValue(r.ISO31661),
			EnglishName: types.StringValue(r.EnglishName),
			Name:        types.StringValue(r.Name),
		})
	}

	regionList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"iso_3166_1":   types.StringType,
			"english_name": types.StringType,
			"name":         types.StringType,
		},
	}, regionElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := RegionsDataSourceModel{
		ID:      types.StringValue("regions"),
		Regions: regionList,
		Total:   types.Int64Value(int64(len(raw))),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
