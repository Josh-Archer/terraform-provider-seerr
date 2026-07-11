package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RequestDataSource{}

type RequestDataSource struct {
	client *APIClient
}

type RequestDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Status       types.Int64  `tfsdk:"status"`
	MediaID      types.Int64  `tfsdk:"media_id"`
	TMDBID       types.Int64  `tfsdk:"tmdb_id"`
	MediaType    types.String `tfsdk:"media_type"`
	Is4K         types.Bool   `tfsdk:"is_4k"`
	UserID       types.Int64  `tfsdk:"user_id"`
	ResponseJSON types.String `tfsdk:"response_json"`
}

func NewRequestDataSource() datasource.DataSource {
	return &RequestDataSource{}
}

func (d *RequestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_request"
}

func (d *RequestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a single Seerr media request by ID via `GET /api/v1/request/{id}`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The request ID.",
				Required:            true,
			},
			"status": schema.Int64Attribute{
				MarkdownDescription: "The status of the request (1: Pending, 2: Approved, 3: Declined).",
				Computed:            true,
			},
			"media_id": schema.Int64Attribute{
				MarkdownDescription: "The Seerr-internal media ID associated with the request (matches `seerr_requests.media_id`).",
				Computed:            true,
			},
			"tmdb_id": schema.Int64Attribute{
				MarkdownDescription: "The TMDB ID of the media associated with the request.",
				Computed:            true,
			},
			"media_type": schema.StringAttribute{
				MarkdownDescription: "The media type (`movie` or `tv`).",
				Computed:            true,
			},
			"is_4k": schema.BoolAttribute{
				MarkdownDescription: "Whether the request is for 4K media.",
				Computed:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the user who made the request.",
				Computed:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body from the API.",
				Computed:            true,
			},
		},
	}
}

func (d *RequestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RequestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RequestDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := d.refreshRequest(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *RequestDataSource) refreshRequest(ctx context.Context, data *RequestDataSourceModel) error {
	id := data.ID.ValueString()
	res, err := d.client.Request(ctx, "GET", "/api/v1/request/"+id, "", nil)
	if err != nil {
		return err
	}
	if res.StatusCode == 404 {
		return fmt.Errorf("request %q not found", id)
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var m map[string]any
	if err := json.Unmarshal(res.Body, &m); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	data.ResponseJSON = types.StringValue(string(res.Body))

	if status, ok := int64ValueFromAny(m["status"]); ok {
		data.Status = types.Int64Value(status)
	}
	if is4k, ok := boolValueFromAny(m["is4k"]); ok {
		data.Is4K = types.BoolValue(is4k)
	}

	if media, ok := m["media"].(map[string]any); ok {
		if mediaID, ok := int64ValueFromAny(media["id"]); ok {
			data.MediaID = types.Int64Value(mediaID)
		}
		if mediaType, ok := stringValueFromAny(media["mediaType"]); ok {
			data.MediaType = types.StringValue(mediaType)
		}
		if tmdbID, ok := int64ValueFromAny(media["tmdbId"]); ok {
			data.TMDBID = types.Int64Value(tmdbID)
		}
	}

	if requestedBy, ok := m["requestedBy"].(map[string]any); ok {
		if userID, ok := int64ValueFromAny(requestedBy["id"]); ok {
			data.UserID = types.Int64Value(userID)
		}
	}

	return nil
}
