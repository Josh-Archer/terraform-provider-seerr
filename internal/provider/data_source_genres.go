package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &GenresDataSource{}

type GenresDataSource struct {
	client *APIClient
}

type GenresDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	MediaType types.String `tfsdk:"media_type"`
	Genres    types.List   `tfsdk:"genres"`
	Total     types.Int64  `tfsdk:"total"`
}

type GenreModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewGenresDataSource() datasource.DataSource {
	return &GenresDataSource{}
}

func (d *GenresDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_genres"
}

func (d *GenresDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch TMDB genre reference lists for movies or TV shows to resolve names to numerical IDs.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier.",
				Computed:            true,
			},
			"media_type": schema.StringAttribute{
				MarkdownDescription: "Media type to fetch genres for (`movie` or `tv`). Defaults to `movie`.",
				Optional:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of genres returned.",
				Computed:            true,
			},
			"genres": schema.ListNestedAttribute{
				MarkdownDescription: "List of genres.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Numerical TMDB genre ID.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Genre name (e.g. Action, Comedy, Drama).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *GenresDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

type rawGenreResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (d *GenresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if d.client == nil {
		resp.Diagnostics.AddError("Unconfigured API Client", "API client was not provided during configuration.")
		return
	}

	var config GenresDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mediaType := "movie"
	if !config.MediaType.IsNull() && !config.MediaType.IsUnknown() && strings.TrimSpace(config.MediaType.ValueString()) != "" {
		mediaType = strings.ToLower(strings.TrimSpace(config.MediaType.ValueString()))
	}
	if mediaType != "movie" && mediaType != "tv" {
		resp.Diagnostics.AddError("Invalid Media Type", "media_type must be either 'movie' or 'tv'.")
		return
	}

	endpoint := fmt.Sprintf("/api/v1/genres/%s", mediaType)
	res, err := d.client.Request(ctx, "GET", endpoint, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var raw []rawGenreResponse
	if err := json.Unmarshal(res.Body, &raw); err != nil {
		resp.Diagnostics.AddError("Error Parsing Genres Response", fmt.Sprintf("Failed to parse response: %s", err))
		return
	}

	genreElements := make([]GenreModel, 0, len(raw))
	for _, g := range raw {
		genreElements = append(genreElements, GenreModel{
			ID:   types.Int64Value(g.ID),
			Name: types.StringValue(g.Name),
		})
	}

	genreList, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":   types.Int64Type,
			"name": types.StringType,
		},
	}, genreElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = types.StringValue(fmt.Sprintf("genres:%s", mediaType))
	config.MediaType = types.StringValue(mediaType)
	config.Genres = genreList
	config.Total = types.Int64Value(int64(len(raw)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
