package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserWatchlistSettingsResource{}
var _ resource.ResourceWithImportState = &UserWatchlistSettingsResource{}

type UserWatchlistSettingsResource struct {
	client *APIClient
}

type UserWatchlistSettingsModel struct {
	ID                  types.String `tfsdk:"id"`
	UserID              types.Int64  `tfsdk:"user_id"`
	WatchlistSyncMovies types.Bool   `tfsdk:"watchlist_sync_movies"`
	WatchlistSyncTv     types.Bool   `tfsdk:"watchlist_sync_tv"`
	ResponseJSON        types.String `tfsdk:"response_json"`
	StatusCode          types.Int64  `tfsdk:"status_code"`
}

func NewUserWatchlistSettingsResource() resource.Resource { return &UserWatchlistSettingsResource{} }

func (r *UserWatchlistSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_watchlist_settings"
}

func (r *UserWatchlistSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr user watchlist sync settings via /api/v1/user/{userId}/settings/watchlist.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the user whose watchlist settings to manage.",
				Required:            true,
			},
			"watchlist_sync_movies": schema.BoolAttribute{
				MarkdownDescription: "Whether to sync movies from the user's Plex watchlist.",
				Required:            true,
			},
			"watchlist_sync_tv": schema.BoolAttribute{
				MarkdownDescription: "Whether to sync TV shows from the user's Plex watchlist.",
				Required:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body from the latest operation.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code from the latest operation.",
				Computed:            true,
			},
		},
	}
}

func (r *UserWatchlistSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func userWatchlistPath(userID int64) string {
	return fmt.Sprintf("/api/v1/user/%d/settings/watchlist", userID)
}

func (r *UserWatchlistSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserWatchlistSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(map[string]any{
		"watchlistSyncMovies": data.WatchlistSyncMovies.ValueBool(),
		"watchlistSyncTv":     data.WatchlistSyncTv.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	apiPath := userWatchlistPath(data.UserID.ValueInt64())
	res, err := r.client.Request(ctx, "POST", apiPath, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", data.UserID.ValueInt64()))
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserWatchlistSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserWatchlistSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := userWatchlistPath(data.UserID.ValueInt64())
	res, err := r.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if res.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
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

func (r *UserWatchlistSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserWatchlistSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(map[string]any{
		"watchlistSyncMovies": data.WatchlistSyncMovies.ValueBool(),
		"watchlistSyncTv":     data.WatchlistSyncTv.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	apiPath := userWatchlistPath(data.UserID.ValueInt64())
	res, err := r.client.Request(ctx, "POST", apiPath, string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", data.UserID.ValueInt64()))
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserWatchlistSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No DELETE route for user watchlist settings; removing from state only.
}

func (r *UserWatchlistSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := requireInt64ID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), id)...)
}
