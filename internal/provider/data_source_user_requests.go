package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserRequestsDataSource{}

type UserRequestsDataSource struct {
	client *APIClient
}

type UserRequestsPageInfoModel struct {
	Page    types.Int64 `tfsdk:"page"`
	Pages   types.Int64 `tfsdk:"pages"`
	Results types.Int64 `tfsdk:"results"`
}

type UserRequestItemModel struct {
	ID        types.String `tfsdk:"id"`
	Status    types.Int64  `tfsdk:"status"`
	MediaID   types.Int64  `tfsdk:"media_id"`
	TMDBID    types.Int64  `tfsdk:"tmdb_id"`
	MediaType types.String `tfsdk:"media_type"`
	Is4K      types.Bool   `tfsdk:"is_4k"`
	CreatedAt types.String `tfsdk:"created_at"`
	UpdatedAt types.String `tfsdk:"updated_at"`
}

type UserRequestsDataSourceModel struct {
	ID       types.String               `tfsdk:"id"`
	UserID   types.Int64                `tfsdk:"user_id"`
	Take     types.Int64                `tfsdk:"take"`
	Skip     types.Int64                `tfsdk:"skip"`
	PageInfo *UserRequestsPageInfoModel `tfsdk:"page_info"`
	Requests []UserRequestItemModel     `tfsdk:"requests"`
}

func NewUserRequestsDataSource() datasource.DataSource {
	return &UserRequestsDataSource{}
}

func (d *UserRequestsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_requests"
}

func (d *UserRequestsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get requests for a specific user from Seerr via `GET /api/v1/user/{userId}/requests`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the user whose requests to retrieve.",
				Required:            true,
			},
			"take": schema.Int64Attribute{
				MarkdownDescription: "Number of requests to retrieve (page size).",
				Optional:            true,
			},
			"skip": schema.Int64Attribute{
				MarkdownDescription: "Number of requests to skip (offset).",
				Optional:            true,
			},
			"page_info": schema.SingleNestedAttribute{
				MarkdownDescription: "Paging details for the returned requests.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"page": schema.Int64Attribute{
						MarkdownDescription: "The current page number.",
						Computed:            true,
					},
					"pages": schema.Int64Attribute{
						MarkdownDescription: "Total number of pages.",
						Computed:            true,
					},
					"results": schema.Int64Attribute{
						MarkdownDescription: "Total number of results across all pages.",
						Computed:            true,
					},
				},
			},
			"requests": schema.ListNestedAttribute{
				MarkdownDescription: "List of requests made by the user.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The request ID.",
							Computed:            true,
						},
						"status": schema.Int64Attribute{
							MarkdownDescription: "Status of the request (1: Pending, 2: Approved, 3: Declined).",
							Computed:            true,
						},
						"media_id": schema.Int64Attribute{
							MarkdownDescription: "Seerr media ID associated with the request.",
							Computed:            true,
						},
						"tmdb_id": schema.Int64Attribute{
							MarkdownDescription: "TMDB ID of the media associated with the request.",
							Computed:            true,
						},
						"media_type": schema.StringAttribute{
							MarkdownDescription: "Media type (`movie` or `tv`).",
							Computed:            true,
						},
						"is_4k": schema.BoolAttribute{
							MarkdownDescription: "Whether the request is for 4K media.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Date and time the request was created in ISO 8601 format.",
							Computed:            true,
						},
						"updated_at": schema.StringAttribute{
							MarkdownDescription: "Date and time the request was last updated in ISO 8601 format.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *UserRequestsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserRequestsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserRequestsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := fmt.Sprintf("/api/v1/user/%d/requests", data.UserID.ValueInt64())
	values := url.Values{}
	if !data.Take.IsNull() && !data.Take.IsUnknown() {
		values.Set("take", strconv.FormatInt(data.Take.ValueInt64(), 10))
	}
	if !data.Skip.IsNull() && !data.Skip.IsUnknown() {
		values.Set("skip", strconv.FormatInt(data.Skip.ValueInt64(), 10))
	}
	if encoded := values.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	res, err := d.client.Request(ctx, "GET", endpoint, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if res.StatusCode == 404 {
		resp.Diagnostics.AddError("User Not Found", fmt.Sprintf("user %d not found", data.UserID.ValueInt64()))
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var payload map[string]any
	if err := json.Unmarshal(res.Body, &payload); err != nil {
		resp.Diagnostics.AddError("Read Failed", "failed to parse API response: "+err.Error())
		return
	}

	pageInfo := UserRequestsPageInfoModel{
		Page:    types.Int64Value(0),
		Pages:   types.Int64Value(0),
		Results: types.Int64Value(0),
	}
	if pi, ok := payload["pageInfo"].(map[string]any); ok {
		if p, ok := int64ValueFromAny(pi["page"]); ok {
			pageInfo.Page = types.Int64Value(p)
		}
		if pages, ok := int64ValueFromAny(pi["pages"]); ok {
			pageInfo.Pages = types.Int64Value(pages)
		}
		if resCount, ok := int64ValueFromAny(pi["results"]); ok {
			pageInfo.Results = types.Int64Value(resCount)
		}
	}
	data.PageInfo = &pageInfo

	requests := make([]UserRequestItemModel, 0)
	if rawResults, ok := payload["results"].([]any); ok {
		for _, item := range rawResults {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			reqItem := UserRequestItemModel{
				CreatedAt: types.StringNull(),
				UpdatedAt: types.StringNull(),
				MediaType: types.StringNull(),
				Is4K:      types.BoolValue(false),
			}

			if idVal, ok := int64ValueFromAny(m["id"]); ok {
				reqItem.ID = types.StringValue(strconv.FormatInt(idVal, 10))
			} else if idStr, ok := stringValueFromAny(m["id"]); ok {
				reqItem.ID = types.StringValue(idStr)
			} else {
				reqItem.ID = types.StringNull()
			}

			if status, ok := int64ValueFromAny(m["status"]); ok {
				reqItem.Status = types.Int64Value(status)
			} else {
				reqItem.Status = types.Int64Null()
			}

			if is4k, ok := boolValueFromAny(m["is4k"]); ok {
				reqItem.Is4K = types.BoolValue(is4k)
			}

			if media, ok := m["media"].(map[string]any); ok {
				if mediaID, ok := int64ValueFromAny(media["id"]); ok {
					reqItem.MediaID = types.Int64Value(mediaID)
				} else {
					reqItem.MediaID = types.Int64Null()
				}
				if mediaType, ok := stringValueFromAny(media["mediaType"]); ok {
					reqItem.MediaType = types.StringValue(mediaType)
				}
				if tmdbID, ok := int64ValueFromAny(media["tmdbId"]); ok {
					reqItem.TMDBID = types.Int64Value(tmdbID)
				} else {
					reqItem.TMDBID = types.Int64Null()
				}
			} else {
				reqItem.MediaID = types.Int64Null()
				reqItem.TMDBID = types.Int64Null()
			}

			if reqItem.MediaType.IsNull() {
				if mediaType, ok := stringValueFromAny(m["mediaType"]); ok {
					reqItem.MediaType = types.StringValue(mediaType)
				} else if mediaType, ok := stringValueFromAny(m["type"]); ok {
					reqItem.MediaType = types.StringValue(mediaType)
				}
			}

			if createdAt, ok := stringValueFromAny(m["createdAt"]); ok {
				reqItem.CreatedAt = types.StringValue(createdAt)
			}
			if updatedAt, ok := stringValueFromAny(m["updatedAt"]); ok {
				reqItem.UpdatedAt = types.StringValue(updatedAt)
			}

			requests = append(requests, reqItem)
		}
	}
	data.Requests = requests

	idStr := fmt.Sprintf("user_%d_requests", data.UserID.ValueInt64())
	if !data.Take.IsNull() && !data.Take.IsUnknown() {
		idStr += fmt.Sprintf("_take_%d", data.Take.ValueInt64())
	}
	if !data.Skip.IsNull() && !data.Skip.IsUnknown() {
		idStr += fmt.Sprintf("_skip_%d", data.Skip.ValueInt64())
	}
	data.ID = types.StringValue(idStr)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
