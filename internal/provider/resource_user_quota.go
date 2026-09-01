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

var _ resource.Resource = &UserQuotaResource{}
var _ resource.ResourceWithImportState = &UserQuotaResource{}

// UserQuotaResource manages per-user movie/TV request quotas.
// It writes to /api/v1/user/{userId}/settings/main (the same endpoint as
// UserWatchlistSettingsResource) but only touches the four quota fields,
// preserving all other main-settings values server-side via a read-merge-write cycle.
type UserQuotaResource struct {
	client *APIClient
}

// UserQuotaModel holds only the quota-related fields exposed by this resource.
// The four "global_*" fields are read-only mirrors of the instance-wide defaults.
type UserQuotaModel struct {
	ID                    types.String `tfsdk:"id"`
	UserID                types.Int64  `tfsdk:"user_id"`
	MovieQuotaLimit       types.Int64  `tfsdk:"movie_quota_limit"`
	MovieQuotaDays        types.Int64  `tfsdk:"movie_quota_days"`
	TvQuotaLimit          types.Int64  `tfsdk:"tv_quota_limit"`
	TvQuotaDays           types.Int64  `tfsdk:"tv_quota_days"`
	GlobalMovieQuotaLimit types.Int64  `tfsdk:"global_movie_quota_limit"`
	GlobalMovieQuotaDays  types.Int64  `tfsdk:"global_movie_quota_days"`
	GlobalTvQuotaLimit    types.Int64  `tfsdk:"global_tv_quota_limit"`
	GlobalTvQuotaDays     types.Int64  `tfsdk:"global_tv_quota_days"`
}

func NewUserQuotaResource() resource.Resource { return &UserQuotaResource{} }

func (r *UserQuotaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_quota"
}

func (r *UserQuotaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage per-user movie and TV request quotas in Seerr via `/api/v1/user/{userId}/settings/main`.\n\n" +
			"Setting a quota to `0` means **unlimited** (the global default applies). " +
			"The `global_*` computed attributes reflect the instance-wide quota defaults for reference.\n\n" +
			"Use `seerr_permission_set` together with this resource to fully declare a user's access tier " +
			"(e.g. *standard_user* vs *power_user*).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the Seerr user whose quotas to manage.",
				Required:            true,
			},
			// --- per-user quota settings (writable) ---
			"movie_quota_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of movie requests allowed within `movie_quota_days`. " +
					"Set to `0` to use the global default (unlimited unless configured globally).",
				Optional: true,
				Computed: true,
			},
			"movie_quota_days": schema.Int64Attribute{
				MarkdownDescription: "Rolling window in days for the movie quota. " +
					"Set to `0` to use the global default.",
				Optional: true,
				Computed: true,
			},
			"tv_quota_limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of TV season requests allowed within `tv_quota_days`. " +
					"Set to `0` to use the global default.",
				Optional: true,
				Computed: true,
			},
			"tv_quota_days": schema.Int64Attribute{
				MarkdownDescription: "Rolling window in days for the TV quota. " +
					"Set to `0` to use the global default.",
				Optional: true,
				Computed: true,
			},
			// --- instance-wide defaults (read-only) ---
			"global_movie_quota_limit": schema.Int64Attribute{
				MarkdownDescription: "Instance-wide default movie quota limit (read-only, from global settings).",
				Computed:            true,
			},
			"global_movie_quota_days": schema.Int64Attribute{
				MarkdownDescription: "Instance-wide default movie quota period in days (read-only, from global settings).",
				Computed:            true,
			},
			"global_tv_quota_limit": schema.Int64Attribute{
				MarkdownDescription: "Instance-wide default TV quota limit (read-only, from global settings).",
				Computed:            true,
			},
			"global_tv_quota_days": schema.Int64Attribute{
				MarkdownDescription: "Instance-wide default TV quota period in days (read-only, from global settings).",
				Computed:            true,
			},
		},
	}
}

func (r *UserQuotaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func userQuotaPath(userID int64) string {
	return fmt.Sprintf("/api/v1/user/%d/settings/main", userID)
}

// Create applies the quota plan then reads back the canonical server state.
func (r *UserQuotaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserQuotaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyQuota(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if err := r.readQuota(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed (read-back)", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes quota state from the API.
func (r *UserQuotaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserQuotaModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := userQuotaPath(data.UserID.ValueInt64())
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

	applyQuotaResponse(&data, res.Body)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update applies the changed quota plan then reads back.
func (r *UserQuotaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserQuotaModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyQuota(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if err := r.readQuota(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed (read-back)", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete removes the resource from state only; Seerr has no DELETE for quota settings.
func (r *UserQuotaResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Quota settings are reset to defaults by the user being deleted, not by a separate endpoint.
	// Removing the resource from state is sufficient.
}

// ImportState imports a user quota resource by user_id (numeric).
func (r *UserQuotaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, err := requireInt64ID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), id)...)
}

// applyQuota performs a read-merge-write so that only the four quota fields are
// modified while all other main-settings are preserved server-side.
func (r *UserQuotaResource) applyQuota(ctx context.Context, data *UserQuotaModel) error {
	apiPath := userQuotaPath(data.UserID.ValueInt64())
	unlock := r.client.LockEndpoint(apiPath)
	defer unlock()

	// Read current settings to preserve non-quota fields.
	res, err := r.client.Request(ctx, "GET", apiPath, "", nil)
	if err != nil {
		return fmt.Errorf("reading current settings: %w", err)
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("reading current settings: status %d: %s", res.StatusCode, string(res.Body))
	}

	var payload map[string]any
	if err := json.Unmarshal(res.Body, &payload); err != nil {
		return fmt.Errorf("parsing current settings: %w", err)
	}

	// Overlay the four quota fields from the plan.
	setOptionalInt64(payload, "movieQuotaLimit", data.MovieQuotaLimit)
	setOptionalInt64(payload, "movieQuotaDays", data.MovieQuotaDays)
	setOptionalInt64(payload, "tvQuotaLimit", data.TvQuotaLimit)
	setOptionalInt64(payload, "tvQuotaDays", data.TvQuotaDays)

	body, _ := json.Marshal(payload)
	writeRes, err := r.client.Request(ctx, "POST", apiPath, string(body), nil)
	if err != nil {
		return fmt.Errorf("writing quota settings: %w", err)
	}
	if !StatusIsOK(writeRes.StatusCode) {
		return fmt.Errorf("writing quota settings: status %d: %s", writeRes.StatusCode, string(writeRes.Body))
	}
	return nil
}

// readQuota fetches the latest main-settings and populates data from the response.
func (r *UserQuotaResource) readQuota(ctx context.Context, data *UserQuotaModel) error {
	res, err := r.client.Request(ctx, "GET", userQuotaPath(data.UserID.ValueInt64()), "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}
	applyQuotaResponse(data, res.Body)
	return nil
}

// applyQuotaResponse maps the JSON response body from /settings/main into data.
func applyQuotaResponse(data *UserQuotaModel, body []byte) {
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", data.UserID.ValueInt64()))

	// Per-user quota fields.
	if v, ok := int64ValueFromAny(decoded["movieQuotaLimit"]); ok {
		data.MovieQuotaLimit = types.Int64Value(v)
	} else {
		data.MovieQuotaLimit = types.Int64Value(0)
	}
	if v, ok := int64ValueFromAny(decoded["movieQuotaDays"]); ok {
		data.MovieQuotaDays = types.Int64Value(v)
	} else {
		data.MovieQuotaDays = types.Int64Value(0)
	}
	if v, ok := int64ValueFromAny(decoded["tvQuotaLimit"]); ok {
		data.TvQuotaLimit = types.Int64Value(v)
	} else {
		data.TvQuotaLimit = types.Int64Value(0)
	}
	if v, ok := int64ValueFromAny(decoded["tvQuotaDays"]); ok {
		data.TvQuotaDays = types.Int64Value(v)
	} else {
		data.TvQuotaDays = types.Int64Value(0)
	}

	// Global (instance-wide) read-only mirrors.
	if v, ok := int64ValueFromAny(decoded["globalMovieQuotaLimit"]); ok {
		data.GlobalMovieQuotaLimit = types.Int64Value(v)
	} else {
		data.GlobalMovieQuotaLimit = types.Int64Null()
	}
	if v, ok := int64ValueFromAny(decoded["globalMovieQuotaDays"]); ok {
		data.GlobalMovieQuotaDays = types.Int64Value(v)
	} else {
		data.GlobalMovieQuotaDays = types.Int64Null()
	}
	if v, ok := int64ValueFromAny(decoded["globalTvQuotaLimit"]); ok {
		data.GlobalTvQuotaLimit = types.Int64Value(v)
	} else {
		data.GlobalTvQuotaLimit = types.Int64Null()
	}
	if v, ok := int64ValueFromAny(decoded["globalTvQuotaDays"]); ok {
		data.GlobalTvQuotaDays = types.Int64Value(v)
	} else {
		data.GlobalTvQuotaDays = types.Int64Null()
	}
}
