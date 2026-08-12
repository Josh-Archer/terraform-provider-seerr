package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WatchlistResource{}
var _ resource.ResourceWithImportState = &WatchlistResource{}

type WatchlistResource struct {
	client *APIClient
}

type WatchlistResourceModel struct {
	ID        types.String `tfsdk:"id"`
	TMDBID    types.Int64  `tfsdk:"tmdb_id"`
	MediaType types.String `tfsdk:"media_type"`
	Title     types.String `tfsdk:"title"`
	Overview  types.String `tfsdk:"overview"`
}

func NewWatchlistResource() resource.Resource {
	return &WatchlistResource{}
}

func (r *WatchlistResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_watchlist"
}

func (r *WatchlistResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage media watchlist items in Seerr.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Resource identifier (format: `tmdb_id:media_type`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tmdb_id": schema.Int64Attribute{
				MarkdownDescription: "The TMDB numeric ID of the movie or TV show.",
				Required:            true,
			},
			"media_type": schema.StringAttribute{
				MarkdownDescription: "Media type (`movie` or `tv`).",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the media item.",
				Computed:            true,
			},
			"overview": schema.StringAttribute{
				MarkdownDescription: "Summary/overview of the media item.",
				Computed:            true,
			},
		},
	}
}

func (r *WatchlistResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *APIClient, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

type watchlistPayload struct {
	TMDBID    int64  `json:"tmdbId"`
	MediaType string `json:"mediaType"`
}

func (r *WatchlistResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan WatchlistResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := watchlistPayload{
		TMDBID:    plan.TMDBID.ValueInt64(),
		MediaType: plan.MediaType.ValueString(),
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error Encoding Watchlist Payload", err.Error())
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/watchlist", string(bodyBytes), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Create") {
		return
	}

	resourceID := fmt.Sprintf("%d:%s", plan.TMDBID.ValueInt64(), plan.MediaType.ValueString())
	plan.ID = types.StringValue(resourceID)

	var item struct {
		Title    string `json:"title"`
		Name     string `json:"name"`
		Overview string `json:"overview"`
	}
	_ = json.Unmarshal(res.Body, &item)
	if item.Title != "" {
		plan.Title = types.StringValue(item.Title)
	} else if item.Name != "" {
		plan.Title = types.StringValue(item.Name)
	} else {
		plan.Title = types.StringValue(fmt.Sprintf("%s (%d)", plan.MediaType.ValueString(), plan.TMDBID.ValueInt64()))
	}
	plan.Overview = types.StringValue(item.Overview)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *WatchlistResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state WatchlistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmdbID := state.TMDBID.ValueInt64()
	mediaType := state.MediaType.ValueString()

	res, err := r.client.Request(ctx, "GET", "/api/v1/watchlist", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if res.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var raw struct {
		Results []struct {
			TMDBID    int64  `json:"tmdbId"`
			MediaType string `json:"mediaType"`
			Title     string `json:"title"`
			Name      string `json:"name"`
			Overview  string `json:"overview"`
		} `json:"results"`
	}

	if err := json.Unmarshal(res.Body, &raw); err != nil {
		resp.Diagnostics.AddError("Error Parsing Watchlist", err.Error())
		return
	}

	found := false
	for _, item := range raw.Results {
		if item.TMDBID == tmdbID && strings.EqualFold(item.MediaType, mediaType) {
			found = true
			if item.Title != "" {
				state.Title = types.StringValue(item.Title)
			} else if item.Name != "" {
				state.Title = types.StringValue(item.Name)
			}
			state.Overview = types.StringValue(item.Overview)
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *WatchlistResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update Not Supported", "Watchlist items cannot be updated in place. Re-create the resource to change media_type or tmdb_id.")
}

func (r *WatchlistResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state WatchlistResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tmdbID := state.TMDBID.ValueInt64()
	endpoint := fmt.Sprintf("/api/v1/watchlist/%d", tmdbID)

	res, err := r.client.Request(ctx, "DELETE", endpoint, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if res.StatusCode != 200 && res.StatusCode != 204 && res.StatusCode != 404 {
		HandleAPIResponse(ctx, res, &resp.Diagnostics, "Delete")
	}
}

func (r *WatchlistResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid Import ID", "Expected import ID format: `tmdb_id:media_type` (e.g. `550:movie`).")
		return
	}

	tmdbID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid TMDB ID", fmt.Sprintf("Could not parse numeric tmdb_id from '%s': %s", parts[0], err))
		return
	}

	mediaType := strings.ToLower(parts[1])
	if mediaType != "movie" && mediaType != "tv" {
		resp.Diagnostics.AddError("Invalid Media Type", "media_type must be either 'movie' or 'tv'.")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("tmdb_id"), tmdbID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("media_type"), mediaType)...)
}
