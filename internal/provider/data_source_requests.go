package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RequestsDataSource{}

type RequestsDataSource struct {
	client *APIClient
}

type RequestSummaryModel struct {
	ID            types.String `tfsdk:"id"`
	Status        types.Int64  `tfsdk:"status"`
	MediaID       types.Int64  `tfsdk:"media_id"`
	MediaType     types.String `tfsdk:"media_type"`
	RequestedByID types.Int64  `tfsdk:"requested_by_id"`
}

type RequestsDataSourceModel struct {
	ID                  types.String          `tfsdk:"id"`
	FilterStatus        types.Int64           `tfsdk:"filter_status"`
	FilterMediaType     types.String          `tfsdk:"filter_media_type"`
	FilterRequestedByID types.Int64           `tfsdk:"filter_requested_by_id"`
	Total               types.Int64           `tfsdk:"total"`
	Requests            []RequestSummaryModel `tfsdk:"requests"`
}

func NewRequestsDataSource() datasource.DataSource {
	return &RequestsDataSource{}
}

func (d *RequestsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_requests"
}

func (d *RequestsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about all existing Seerr requests.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"filter_status": schema.Int64Attribute{
				MarkdownDescription: "Filter by request status (1=Pending, 2=Approved, 3=Declined).",
				Optional:            true,
			},
			"filter_media_type": schema.StringAttribute{
				MarkdownDescription: "Filter by media type (movie or tv).",
				Optional:            true,
			},
			"filter_requested_by_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by requesting user ID.",
				Optional:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of returned requests.",
				Computed:            true,
			},
			"requests": schema.ListNestedAttribute{
				MarkdownDescription: "List of requests.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the request.",
							Computed:            true,
						},
						"status": schema.Int64Attribute{
							MarkdownDescription: "The status of the request (1: Pending, 2: Approved, 3: Declined).",
							Computed:            true,
						},
						"media_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the media associated with the request.",
							Computed:            true,
						},
						"media_type": schema.StringAttribute{
							MarkdownDescription: "The type of media (movie or tv).",
							Computed:            true,
						},
						"requested_by_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the user who made the request.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *RequestsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RequestsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RequestsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	results, err := fetchAllPaginatedResults(ctx, d.client, "/api/v1/request", defaultPaginationPageSize)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	var filteredRequests []RequestSummaryModel

	for _, u := range results {
		request := RequestSummaryModel{}

		idRaw := u["id"]
		switch v := idRaw.(type) {
		case float64:
			request.ID = types.StringValue(fmt.Sprintf("%.0f", v))
		case string:
			request.ID = types.StringValue(v)
		}

		if s, ok := u["status"].(float64); ok {
			request.Status = types.Int64Value(int64(s))
		}
		if t, ok := u["type"].(string); ok {
			request.MediaType = types.StringValue(t)
		}

		if mediaRaw, ok := u["media"].(map[string]any); ok {
			if mediaId, ok := mediaRaw["id"].(float64); ok {
				request.MediaID = types.Int64Value(int64(mediaId))
			}
		}

		if requestedByRaw, ok := u["requestedBy"].(map[string]any); ok {
			if requestedById, ok := requestedByRaw["id"].(float64); ok {
				request.RequestedByID = types.Int64Value(int64(requestedById))
			}
		}

		// Apply filters
		if !data.FilterStatus.IsNull() && !data.FilterStatus.IsUnknown() {
			if request.Status.ValueInt64() != data.FilterStatus.ValueInt64() {
				continue
			}
		}
		if !data.FilterMediaType.IsNull() && !data.FilterMediaType.IsUnknown() {
			if request.MediaType.ValueString() != data.FilterMediaType.ValueString() {
				continue
			}
		}
		if !data.FilterRequestedByID.IsNull() && !data.FilterRequestedByID.IsUnknown() {
			if request.RequestedByID.ValueInt64() != data.FilterRequestedByID.ValueInt64() {
				continue
			}
		}

		filteredRequests = append(filteredRequests, request)
	}

	if filteredRequests == nil {
		filteredRequests = []RequestSummaryModel{}
	}

	data.Requests = filteredRequests
	data.Total = types.Int64Value(int64(len(filteredRequests)))

	idStr := "all_requests"
	if !data.FilterStatus.IsNull() {
		idStr += fmt.Sprintf("_status_%d", data.FilterStatus.ValueInt64())
	}
	if !data.FilterMediaType.IsNull() {
		idStr += fmt.Sprintf("_type_%s", data.FilterMediaType.ValueString())
	}
	if !data.FilterRequestedByID.IsNull() {
		idStr += fmt.Sprintf("_user_%d", data.FilterRequestedByID.ValueInt64())
	}
	data.ID = types.StringValue(idStr)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
