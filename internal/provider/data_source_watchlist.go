package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &WatchlistDataSource{}

type WatchlistDataSource struct {
	client *APIClient
}

type WatchlistDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	UserID    types.Int64  `tfsdk:"user_id"`
	Total     types.Int64  `tfsdk:"total"`
	Watchlist types.List   `tfsdk:"watchlist"`
}

type WatchlistItemModel struct {
	TMDBID    types.Int64  `tfsdk:"tmdb_id"`
	MediaType types.String `tfsdk:"media_type"`
	Title     types.String `tfsdk:"title"`
	Overview  types.String `tfsdk:"overview"`
}

func NewWatchlistDataSource() datasource.DataSource {
	return &WatchlistDataSource{}
}

func (d *WatchlistDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_watchlist"
}

func (d *WatchlistDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch all pages of watchlist items for the authenticated user or a specific user in Seerr via `/api/v1/user/{id}/watchlist`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier.",
				Computed:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "Optional target user ID. Omit to fetch authenticated user's watchlist.",
				Optional:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of watchlist items reported by Seerr (`totalResults`).",
				Computed:            true,
			},
			"watchlist": schema.ListNestedAttribute{
				MarkdownDescription: "List of watchlist items.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"tmdb_id": schema.Int64Attribute{
							MarkdownDescription: "TMDB numeric media ID.",
							Computed:            true,
						},
						"media_type": schema.StringAttribute{
							MarkdownDescription: "Media type (`movie` or `tv`).",
							Computed:            true,
						},
						"title": schema.StringAttribute{
							MarkdownDescription: "Title of media item.",
							Computed:            true,
						},
						"overview": schema.StringAttribute{
							MarkdownDescription: "Media summary/overview.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *WatchlistDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WatchlistDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config WatchlistDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	userID := int64(0)
	if !config.UserID.IsNull() && !config.UserID.IsUnknown() {
		userID = config.UserID.ValueInt64()
		if userID <= 0 {
			resp.Diagnostics.AddError("Read Failed", "user_id must be a positive integer")
			return
		}
	} else {
		var err error
		userID, err = authenticatedUserID(ctx, d.client)
		if err != nil {
			resp.Diagnostics.AddError("Read Failed", err.Error())
			return
		}
	}
	items, total, err := fetchWatchlist(ctx, d.client, userID)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	watchlist := make([]WatchlistItemModel, 0, len(items))
	for _, r := range items {
		title := r.Title
		if title == "" {
			title = r.Name
		}
		watchlist = append(watchlist, WatchlistItemModel{
			TMDBID:    types.Int64Value(*r.TMDBID),
			MediaType: types.StringValue(r.MediaType),
			Title:     types.StringValue(title),
			Overview:  types.StringValue(r.Overview),
		})
	}

	watchlistVal, diags := types.ListValueFrom(ctx, types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"tmdb_id":    types.Int64Type,
			"media_type": types.StringType,
			"title":      types.StringType,
			"overview":   types.StringType,
		},
	}, watchlist)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.ID = types.StringValue(fmt.Sprintf("/api/v1/user/%d/watchlist", userID))
	config.Watchlist = watchlistVal
	config.Total = types.Int64Value(total)

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
