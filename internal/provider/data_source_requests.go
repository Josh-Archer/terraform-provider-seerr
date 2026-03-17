package provider

import (
	"context"
	"encoding/json"
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
	RequestedByID types.Int64  `tfsdk:"requested_by_id"`
}

type RequestsDataSourceModel struct {
	ID       types.String          `tfsdk:"id"`
	Requests []RequestSummaryModel `tfsdk:"requests"`
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

func (d *RequestsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RequestsDataSourceModel

	// Fetch requests
	res, err := d.client.Request(ctx, "GET", "/api/v1/request?take=1000", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var parsedResponse struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(res.Body, &parsedResponse); err != nil {
		resp.Diagnostics.AddError("Read Failed", "Failed to parse API response: "+err.Error())
		return
	}

	for _, u := range parsedResponse.Results {
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

		data.Requests = append(data.Requests, request)
	}

	data.ID = types.StringValue("all_requests")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
