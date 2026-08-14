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

var _ datasource.DataSource = &LanguagesDataSource{}

type LanguagesDataSource struct {
	client *APIClient
}

type LanguagesDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Languages types.List   `tfsdk:"languages"`
	Total     types.Int64  `tfsdk:"total"`
}

type LanguageModel struct {
	ISO6391     types.String `tfsdk:"iso_639_1"`
	EnglishName types.String `tfsdk:"english_name"`
	Name        types.String `tfsdk:"name"`
}

func NewLanguagesDataSource() datasource.DataSource {
	return &LanguagesDataSource{}
}

func (d *LanguagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_languages"
}

func (d *LanguagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch supported ISO language codes from TMDB for language-filtered discover sliders.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier.",
				Computed:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of languages returned.",
				Computed:            true,
			},
			"languages": schema.ListNestedAttribute{
				MarkdownDescription: "List of languages.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"iso_639_1": schema.StringAttribute{
							MarkdownDescription: "ISO 639-1 language code (e.g. `en`, `es`, `fr`).",
							Computed:            true,
						},
						"english_name": schema.StringAttribute{
							MarkdownDescription: "English name of the language.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Native name of the language.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *LanguagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

type rawLanguageResponse struct {
	ISO6391     string `json:"iso_639_1"`
	EnglishName string `json:"english_name"`
	Name        string `json:"name"`
}

func (d *LanguagesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured API Client", "API client was not provided during configuration.")
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/languages", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var raw []rawLanguageResponse
	if err := json.Unmarshal(res.Body, &raw); err != nil {
		resp.Diagnostics.AddError("Error Parsing Languages Response", fmt.Sprintf("Failed to parse response: %s", err))
		return
	}

	langElements := make([]LanguageModel, 0, len(raw))
	for _, l := range raw {
		langElements = append(langElements, LanguageModel{
			ISO6391:     types.StringValue(l.ISO6391),
			EnglishName: types.StringValue(l.EnglishName),
			Name:        types.StringValue(l.Name),
		})
	}

	langList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"iso_639_1":    types.StringType,
			"english_name": types.StringType,
			"name":         types.StringType,
		},
	}, langElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state := LanguagesDataSourceModel{
		ID:        types.StringValue("languages"),
		Languages: langList,
		Total:     types.Int64Value(int64(len(raw))),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
