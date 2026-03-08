package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &SonarrServerResource{}
var _ resource.ResourceWithImportState = &SonarrServerResource{}

type SonarrServerResource struct {
	client *APIClient
}

type SonarrServerModel struct {
	ID                   types.String `tfsdk:"id"`
	ServerID             types.Int64  `tfsdk:"server_id"`
	Name                 types.String `tfsdk:"name"`
	URL                  types.String `tfsdk:"url"`
	Hostname             types.String `tfsdk:"hostname"`
	Port                 types.Int64  `tfsdk:"port"`
	APIKey               types.String `tfsdk:"api_key"`
	UseSSL               types.Bool   `tfsdk:"use_ssl"`
	BaseURL              types.String `tfsdk:"base_url"`
	QualityProfileID     types.Int64  `tfsdk:"quality_profile_id"`
	QualityProfileName   types.String `tfsdk:"quality_profile_name"`
	ActiveDirectory      types.String `tfsdk:"active_directory"`
	ActiveAnimeDirectory types.String `tfsdk:"active_anime_directory"`
	Tags                 types.List   `tfsdk:"tags"`
	AnimeTags            types.List   `tfsdk:"anime_tags"`
	Is4K                 types.Bool   `tfsdk:"is_4k"`
	IsDefault            types.Bool   `tfsdk:"is_default"`
	EnableScan           types.Bool   `tfsdk:"enable_scan"`
	EnableSeasonFolders  types.Bool   `tfsdk:"enable_season_folders"`
	SyncEnabled          types.Bool   `tfsdk:"sync_enabled"`
	PreventSearch        types.Bool   `tfsdk:"prevent_search"`
	TagRequestsWithUser  types.Bool   `tfsdk:"tag_requests_with_user"`
	ExtraPayloadJSON     types.String `tfsdk:"extra_payload_json"`
	ResponseJSON         types.String `tfsdk:"response_json"`
}

func NewSonarrServerResource() resource.Resource { return &SonarrServerResource{} }

func (r *SonarrServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sonarr_server"
}

func (r *SonarrServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr Sonarr server settings via /api/v1/settings/sonarr.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("Sonarr"),
			},
			"url": schema.StringAttribute{Optional: true},
			"hostname": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("sonarr-service"),
			},
			"port": schema.Int64Attribute{Optional: true, Computed: true},
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"use_ssl": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"base_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(""),
			},
			"quality_profile_id": schema.Int64Attribute{Required: true},
			"quality_profile_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"active_directory":       schema.StringAttribute{Required: true},
			"active_anime_directory": schema.StringAttribute{Optional: true},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"anime_tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"is_4k": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"is_default": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"enable_scan": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"enable_season_folders": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"sync_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"prevent_search": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"tag_requests_with_user": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"extra_payload_json": schema.StringAttribute{Optional: true},
			"response_json":      schema.StringAttribute{Computed: true},
		},
	}
}

func (r *SonarrServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseSonarrURLIntoModel(data *SonarrServerModel) {
	if data.URL.IsNull() || data.URL.IsUnknown() || strings.TrimSpace(data.URL.ValueString()) == "" {
		if data.Port.IsNull() || data.Port.IsUnknown() {
			data.Port = types.Int64Value(8989)
		}
		return
	}
	u, err := url.Parse(data.URL.ValueString())
	if err != nil {
		return
	}
	if u.Hostname() != "" {
		data.Hostname = types.StringValue(u.Hostname())
	}
	if u.Port() != "" {
		if p, err := strconv.ParseInt(u.Port(), 10, 64); err == nil {
			data.Port = types.Int64Value(p)
		}
	} else if data.Port.IsNull() || data.Port.IsUnknown() {
		if u.Scheme == "https" {
			data.Port = types.Int64Value(443)
		} else {
			data.Port = types.Int64Value(80)
		}
	}
	if u.Path != "" && u.Path != "/" {
		data.BaseURL = types.StringValue(u.Path)
	}
	if u.Scheme == "https" {
		data.UseSSL = types.BoolValue(true)
	}
}

func listInt64(ctx context.Context, l types.List) ([]int64, error) {
	if l.IsNull() || l.IsUnknown() {
		return []int64{}, nil
	}
	var vals []int64
	if diags := l.ElementsAs(ctx, &vals, false); diags.HasError() {
		return nil, fmt.Errorf("invalid list")
	}
	return vals, nil
}

func (r *SonarrServerResource) payload(ctx context.Context, data SonarrServerModel) (SonarrServerModel, string, error) {
	parseSonarrURLIntoModel(&data)
	tags, err := listInt64(ctx, data.Tags)
	if err != nil {
		return data, "", err
	}
	animeTags, err := listInt64(ctx, data.AnimeTags)
	if err != nil {
		return data, "", err
	}
	animeDir := data.ActiveAnimeDirectory.ValueString()
	if animeDir == "" {
		animeDir = data.ActiveDirectory.ValueString()
	}
	profileName := ""
	if !data.QualityProfileName.IsNull() && !data.QualityProfileName.IsUnknown() {
		profileName = strings.TrimSpace(data.QualityProfileName.ValueString())
	}
	if profileName == "" {
		profileID := data.QualityProfileID.ValueInt64()
		profile, lookupErr := findArrProfile(
			ctx,
			data.URL.ValueString(),
			data.Hostname.ValueString(),
			data.Port.ValueInt64(),
			data.UseSSL.ValueBool(),
			data.BaseURL.ValueString(),
			data.APIKey.ValueString(),
			&profileID,
			nil,
		)
		if lookupErr != nil {
			return data, "", fmt.Errorf("resolve quality_profile_name: %w", lookupErr)
		}
		profileName = profile.Name
	}
	data.QualityProfileName = types.StringValue(profileName)
	base := map[string]any{
		"name":                 data.Name.ValueString(),
		"hostname":             data.Hostname.ValueString(),
		"port":                 data.Port.ValueInt64(),
		"apiKey":               data.APIKey.ValueString(),
		"useSsl":               data.UseSSL.ValueBool(),
		"baseUrl":              data.BaseURL.ValueString(),
		"activeProfileId":      data.QualityProfileID.ValueInt64(),
		"activeProfileName":    profileName,
		"activeDirectory":      data.ActiveDirectory.ValueString(),
		"activeAnimeDirectory": animeDir,
		"tags":                 tags,
		"animeTags":            animeTags,
		"is4k":                 data.Is4K.ValueBool(),
		"isDefault":            data.IsDefault.ValueBool(),
		"enableScan":           data.EnableScan.ValueBool(),
		"enableSeasonFolders":  data.EnableSeasonFolders.ValueBool(),
		"syncEnabled":          data.SyncEnabled.ValueBool(),
		"preventSearch":        data.PreventSearch.ValueBool(),
		"tagRequests":          data.TagRequestsWithUser.ValueBool(),
	}
	merged, err := mergeJSON(base, data.ExtraPayloadJSON.ValueString())
	if err != nil {
		return data, "", err
	}
	b, err := json.Marshal(merged)
	if err != nil {
		return data, "", err
	}
	return data, string(b), nil
}

func (r *SonarrServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SonarrServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	normalizedData, body, err := r.payload(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	data = normalizedData
	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/sonarr", body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}
	id, ok := ExtractIDFromJSON(res.Body)
	if !ok {
		resp.Diagnostics.AddError("Create Failed", "response did not include Sonarr server id")
		return
	}
	parsed, _ := requireInt64ID(id)
	data.ServerID = types.Int64Value(parsed)
	data.ID = types.StringValue(id)
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SonarrServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SonarrServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/sonarr", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}
	item, found, err := findByIDInJSONArray(res.Body, data.ServerID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}
	data.ResponseJSON = types.StringValue(string(item))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SonarrServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state SonarrServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var data SonarrServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID, data.ServerID = normalizeServerIdentity(data.ID, state.ID, data.ServerID, state.ServerID)
	normalizedData, body, err := r.payload(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	data = normalizedData
	path := fmt.Sprintf("/api/v1/settings/sonarr/%d", data.ServerID.ValueInt64())
	res, err := r.client.Request(ctx, "PUT", path, body, nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}
	data.ID = types.StringValue(fmt.Sprintf("%d", data.ServerID.ValueInt64()))
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SonarrServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SonarrServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := fmt.Sprintf("/api/v1/settings/sonarr/%d", data.ServerID.ValueInt64())
	res, err := r.client.Request(ctx, "DELETE", path, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Delete Failed", err.Error())
		return
	}
	if res.StatusCode == 404 {
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Delete Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
	}
}

func (r *SonarrServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	id, err := requireInt64ID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_id"), id)...)
}
