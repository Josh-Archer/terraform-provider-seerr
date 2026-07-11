package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &MediaItemDataSource{}

type MediaItemDataSource struct {
	client *APIClient
}

type MediaItemDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	TMDBID       types.Int64  `tfsdk:"tmdb_id"`
	TVDBID       types.Int64  `tfsdk:"tvdb_id"`
	MediaType    types.String `tfsdk:"media_type"`
	Status       types.Int64  `tfsdk:"status"`
	Status4k     types.Int64  `tfsdk:"status_4k"`
	ResponseJSON types.String `tfsdk:"response_json"`
}

func NewMediaItemDataSource() datasource.DataSource {
	return &MediaItemDataSource{}
}

func (d *MediaItemDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_media_item"
}

func (d *MediaItemDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a single Seerr media record by ID via `GET /api/v1/media/{id}`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The Seerr-internal media ID.",
				Required:            true,
			},
			"tmdb_id": schema.Int64Attribute{
				MarkdownDescription: "The TMDB ID of the media.",
				Computed:            true,
			},
			"tvdb_id": schema.Int64Attribute{
				MarkdownDescription: "The TVDB ID of the media, when present.",
				Computed:            true,
			},
			"media_type": schema.StringAttribute{
				MarkdownDescription: "The media type (`movie` or `tv`).",
				Computed:            true,
			},
			"status": schema.Int64Attribute{
				MarkdownDescription: "The media availability status.",
				Computed:            true,
			},
			"status_4k": schema.Int64Attribute{
				MarkdownDescription: "The 4K media availability status.",
				Computed:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body from the API.",
				Computed:            true,
			},
		},
	}
}

func (d *MediaItemDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MediaItemDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MediaItemDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := d.refreshMediaItem(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *MediaItemDataSource) refreshMediaItem(ctx context.Context, data *MediaItemDataSourceModel) error {
	id := data.ID.ValueString()
	res, err := d.client.Request(ctx, "GET", "/api/v1/media/"+id, "", nil)
	if err != nil {
		return err
	}
	if res.StatusCode == 404 {
		return fmt.Errorf("media item %q not found", id)
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var m map[string]any
	if err := json.Unmarshal(res.Body, &m); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	data.ResponseJSON = types.StringValue(string(res.Body))

	if tmdbID, ok := int64ValueFromAny(m["tmdbId"]); ok {
		data.TMDBID = types.Int64Value(tmdbID)
	}
	if tvdbID, ok := int64ValueFromAny(m["tvdbId"]); ok {
		data.TVDBID = types.Int64Value(tvdbID)
	} else {
		data.TVDBID = types.Int64Null()
	}
	if mediaType, ok := stringValueFromAny(m["mediaType"]); ok {
		data.MediaType = types.StringValue(mediaType)
	}
	if status, ok := int64ValueFromAny(m["status"]); ok {
		data.Status = types.Int64Value(status)
	}
	if status4k, ok := int64ValueFromAny(m["status4k"]); ok {
		data.Status4k = types.Int64Value(status4k)
	}

	return nil
}
