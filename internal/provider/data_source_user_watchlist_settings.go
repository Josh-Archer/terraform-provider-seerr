package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserWatchlistSettingsDataSource{}

type UserWatchlistSettingsDataSource struct {
	client *APIClient
}

type UserWatchlistSettingsDataSourceModel struct {
	UserID              types.Int64  `tfsdk:"user_id"`
	WatchlistSyncMovies types.Bool   `tfsdk:"watchlist_sync_movies"`
	WatchlistSyncTv     types.Bool   `tfsdk:"watchlist_sync_tv"`
	ResponseJSON        types.String `tfsdk:"response_json"`
	StatusCode          types.Int64  `tfsdk:"status_code"`
}

func NewUserWatchlistSettingsDataSource() datasource.DataSource {
	return &UserWatchlistSettingsDataSource{}
}

func (d *UserWatchlistSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_watchlist_settings"
}

func (d *UserWatchlistSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Seerr user watchlist sync settings via /api/v1/user/{userId}/settings/main.",
		Attributes: map[string]schema.Attribute{
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the user to look up.",
				Required:            true,
			},
			"watchlist_sync_movies": schema.BoolAttribute{
				MarkdownDescription: "Whether movies are synced from the user's Plex watchlist.",
				Computed:            true,
			},
			"watchlist_sync_tv": schema.BoolAttribute{
				MarkdownDescription: "Whether TV shows are synced from the user's Plex watchlist.",
				Computed:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code.",
				Computed:            true,
			},
		},
	}
}

func (d *UserWatchlistSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserWatchlistSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserWatchlistSettingsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := userWatchlistPath(data.UserID.ValueInt64())
	res, err := d.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err == nil {
		if raw, ok := decoded["watchlistSyncMovies"]; ok {
			if v, ok := raw.(bool); ok {
				data.WatchlistSyncMovies = types.BoolValue(v)
			}
		}
		if raw, ok := decoded["watchlistSyncTv"]; ok {
			if v, ok := raw.(bool); ok {
				data.WatchlistSyncTv = types.BoolValue(v)
			}
		}
	}

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
