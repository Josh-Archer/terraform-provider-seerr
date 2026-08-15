package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RadarrServerResource{}
var _ resource.ResourceWithImportState = &RadarrServerResource{}

type RadarrServerResource struct {
	client *APIClient
}

type RadarrServerModel struct {
	ID                  types.String `tfsdk:"id"`
	ServerID            types.Int64  `tfsdk:"server_id"`
	Name                types.String `tfsdk:"name"`
	URL                 types.String `tfsdk:"url"`
	Hostname            types.String `tfsdk:"hostname"`
	Port                types.Int64  `tfsdk:"port"`
	APIKey              types.String `tfsdk:"api_key"`
	UseSSL              types.Bool   `tfsdk:"use_ssl"`
	BaseURL             types.String `tfsdk:"base_url"`
	QualityProfileID    types.Int64  `tfsdk:"quality_profile_id"`
	QualityProfileName  types.String `tfsdk:"quality_profile_name"`
	ActiveDirectory     types.String `tfsdk:"active_directory"`
	Is4K                types.Bool   `tfsdk:"is_4k"`
	MinimumAvailability types.String `tfsdk:"minimum_availability"`
	Tags                types.List   `tfsdk:"tags"`
	IsDefault           types.Bool   `tfsdk:"is_default"`
	EnableScan          types.Bool   `tfsdk:"enable_scan"`
	SyncEnabled         types.Bool   `tfsdk:"sync_enabled"`
	PreventSearch       types.Bool   `tfsdk:"prevent_search"`
	TagRequestsWithUser types.Bool   `tfsdk:"tag_requests_with_user"`
	ExtraPayloadJSON    types.String `tfsdk:"extra_payload_json"`
}

func NewRadarrServerResource() resource.Resource { return &RadarrServerResource{} }

func (r *RadarrServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radarr_server"
}

func (r *RadarrServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr Radarr server settings via /api/v1/settings/radarr.",
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("hostname")),
					stringvalidator.ConflictsWith(path.MatchRoot("port")),
					stringvalidator.RegexMatches(urlRegex(), "must be a valid HTTP or HTTPS URL (e.g., http://localhost:7878)"),
				},
			},
			"hostname": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("url")),
				},
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int64{
					int64validator.ConflictsWith(path.MatchRoot("url")),
					int64validator.Between(1, 65535),
				},
			},
			"api_key": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"use_ssl": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"base_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"quality_profile_id": schema.Int64Attribute{Required: true},
			"quality_profile_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active_directory": schema.StringAttribute{Required: true},
			"is_4k": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"minimum_availability": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"is_default": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"enable_scan": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"sync_enabled": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"prevent_search": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"tag_requests_with_user": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"extra_payload_json": schema.StringAttribute{Optional: true},
		},
	}
}

func (r *RadarrServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func parseURLIntoModel(data *RadarrServerModel) {
	if data.URL.IsNull() || data.URL.IsUnknown() || strings.TrimSpace(data.URL.ValueString()) == "" {
		if data.Port.IsNull() || data.Port.IsUnknown() {
			data.Port = types.Int64Value(7878)
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

func (r *RadarrServerResource) payload(ctx context.Context, data RadarrServerModel) (RadarrServerModel, string, error) {
	parseURLIntoModel(&data)
	tags, err := listInt64(ctx, data.Tags)
	if err != nil {
		return data, "", err
	}

	name := "Radarr"
	if !data.Name.IsNull() && !data.Name.IsUnknown() && strings.TrimSpace(data.Name.ValueString()) != "" {
		name = strings.TrimSpace(data.Name.ValueString())
	}
	data.Name = types.StringValue(name)

	hostname := "radarr-service"
	if !data.Hostname.IsNull() && !data.Hostname.IsUnknown() && strings.TrimSpace(data.Hostname.ValueString()) != "" {
		hostname = strings.TrimSpace(data.Hostname.ValueString())
	}
	data.Hostname = types.StringValue(hostname)

	port := int64(7878)
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		port = data.Port.ValueInt64()
	}
	data.Port = types.Int64Value(port)

	useSSL := false
	if !data.UseSSL.IsNull() && !data.UseSSL.IsUnknown() {
		useSSL = data.UseSSL.ValueBool()
	}
	data.UseSSL = types.BoolValue(useSSL)

	baseURL := ""
	if !data.BaseURL.IsNull() && !data.BaseURL.IsUnknown() {
		baseURL = data.BaseURL.ValueString()
	}
	data.BaseURL = types.StringValue(baseURL)

	is4K := false
	if !data.Is4K.IsNull() && !data.Is4K.IsUnknown() {
		is4K = data.Is4K.ValueBool()
	}
	data.Is4K = types.BoolValue(is4K)

	minAvail := "announced"
	if !data.MinimumAvailability.IsNull() && !data.MinimumAvailability.IsUnknown() && strings.TrimSpace(data.MinimumAvailability.ValueString()) != "" {
		minAvail = strings.TrimSpace(data.MinimumAvailability.ValueString())
	}
	data.MinimumAvailability = types.StringValue(minAvail)

	isDefault := true
	if !data.IsDefault.IsNull() && !data.IsDefault.IsUnknown() {
		isDefault = data.IsDefault.ValueBool()
	}
	data.IsDefault = types.BoolValue(isDefault)

	syncEnabled := true
	if !data.SyncEnabled.IsNull() && !data.SyncEnabled.IsUnknown() {
		syncEnabled = data.SyncEnabled.ValueBool()
	}
	data.SyncEnabled = types.BoolValue(syncEnabled)

	preventSearch := false
	if !data.PreventSearch.IsNull() && !data.PreventSearch.IsUnknown() {
		preventSearch = data.PreventSearch.ValueBool()
	}
	data.PreventSearch = types.BoolValue(preventSearch)

	tagRequests := true
	if !data.TagRequestsWithUser.IsNull() && !data.TagRequestsWithUser.IsUnknown() {
		tagRequests = data.TagRequestsWithUser.ValueBool()
	}
	data.TagRequestsWithUser = types.BoolValue(tagRequests)

	profileName := ""
	if !data.QualityProfileName.IsNull() && !data.QualityProfileName.IsUnknown() {
		profileName = strings.TrimSpace(data.QualityProfileName.ValueString())
	}
	if profileName == "" {
		profileID := data.QualityProfileID.ValueInt64()
		profile, lookupErr := findArrProfile(
			ctx,
			data.URL.ValueString(),
			hostname,
			port,
			useSSL,
			baseURL,
			data.APIKey.ValueString(),
			r.client.Timeout(),
			&profileID,
			nil,
		)
		if lookupErr != nil {
			return data, "", fmt.Errorf("resolve quality_profile_name: %w", lookupErr)
		}
		profileName = profile.Name
	}
	data.QualityProfileName = types.StringValue(profileName)

	// Validate connectivity to Radarr
	if err := ValidateArrConnectivity(
		ctx,
		data.URL.ValueString(),
		hostname,
		port,
		useSSL,
		baseURL,
		data.APIKey.ValueString(),
		r.client.Timeout(),
	); err != nil {
		return data, "", fmt.Errorf("validate connectivity: %w", err)
	}

	base := map[string]any{
		"name":                name,
		"hostname":            hostname,
		"port":                port,
		"apiKey":              data.APIKey.ValueString(),
		"useSsl":              useSSL,
		"baseUrl":             baseURL,
		"activeProfileId":     data.QualityProfileID.ValueInt64(),
		"activeProfileName":   profileName,
		"activeDirectory":     data.ActiveDirectory.ValueString(),
		"is4k":                is4K,
		"minimumAvailability": minAvail,
		"tags":                tags,
		"isDefault":           isDefault,
		"syncEnabled":         syncEnabled,
		"preventSearch":       preventSearch,
		"tagRequests":         tagRequests,
	}
	setOptionalBool(base, "enableScan", data.EnableScan)
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

func (r *RadarrServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RadarrServerModel
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
	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/radarr", body, nil)
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
		resp.Diagnostics.AddError("Create Failed", "response did not include Radarr server id")
		return
	}
	parsed, _ := requireInt64ID(id)
	data.ServerID = types.Int64Value(parsed)
	data.ID = types.StringValue(id)
	if err := readRadarrStateFromJSON(ctx, res.Body, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("read state after create: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readRadarrStateFromJSON populates all API-sourced fields of data from the
// JSON object representing a single Radarr server entry returned by Overseerr.
// Fields that are user-supplied and never echoed by the API (url and
// extra_payload_json) are intentionally left untouched. api_key is preserved
// from state when it is already set, but data source reads can still populate
// it when the API returns a value.
func readRadarrStateFromJSON(ctx context.Context, item []byte, data *RadarrServerModel) error {
	var m map[string]any
	if err := json.Unmarshal(item, &m); err != nil {
		return fmt.Errorf("parse radarr server response: %w", err)
	}

	if val, ok := m["name"]; ok {
		if v, ok := val.(string); ok {
			data.Name = types.StringValue(v)
		} else if val == nil {
			data.Name = types.StringNull()
		}
	}
	if val, ok := m["hostname"]; ok {
		if v, ok := val.(string); ok {
			data.Hostname = types.StringValue(v)
		} else if val == nil {
			data.Hostname = types.StringNull()
		}
	}
	if val, ok := m["port"]; ok {
		if v, ok := val.(float64); ok {
			data.Port = types.Int64Value(int64(v))
		} else if val == nil {
			data.Port = types.Int64Null()
		}
	}
	if val, ok := m["useSsl"]; ok {
		if v, ok := val.(bool); ok {
			data.UseSSL = types.BoolValue(v)
		} else if val == nil {
			data.UseSSL = types.BoolNull()
		}
	}
	if val, ok := m["baseUrl"]; ok {
		if v, ok := val.(string); ok {
			data.BaseURL = types.StringValue(v)
		} else if val == nil {
			data.BaseURL = types.StringNull()
		}
	}
	if val, ok := m["activeProfileId"]; ok {
		if v, ok := val.(float64); ok {
			data.QualityProfileID = types.Int64Value(int64(v))
		} else if val == nil {
			data.QualityProfileID = types.Int64Null()
		}
	}
	if val, ok := m["activeProfileName"]; ok {
		if v, ok := val.(string); ok {
			data.QualityProfileName = types.StringValue(v)
		} else if val == nil {
			data.QualityProfileName = types.StringNull()
		}
	}
	if val, ok := m["activeDirectory"]; ok {
		if v, ok := val.(string); ok {
			data.ActiveDirectory = types.StringValue(v)
		} else if val == nil {
			data.ActiveDirectory = types.StringNull()
		}
	}
	if val, ok := m["is4k"]; ok {
		if v, ok := val.(bool); ok {
			data.Is4K = types.BoolValue(v)
		} else if val == nil {
			data.Is4K = types.BoolNull()
		}
	}
	if val, ok := m["minimumAvailability"]; ok {
		if v, ok := val.(string); ok {
			data.MinimumAvailability = types.StringValue(v)
		} else if val == nil {
			data.MinimumAvailability = types.StringNull()
		}
	}
	if val, ok := m["isDefault"]; ok {
		if v, ok := val.(bool); ok {
			data.IsDefault = types.BoolValue(v)
		} else if val == nil {
			data.IsDefault = types.BoolNull()
		}
	}
	if val, ok := m["enableScan"]; ok {
		if v, ok := val.(bool); ok {
			data.EnableScan = types.BoolValue(v)
		} else if val == nil {
			data.EnableScan = types.BoolNull()
		}
	}
	if val, ok := m["syncEnabled"]; ok {
		if v, ok := val.(bool); ok {
			data.SyncEnabled = types.BoolValue(v)
		} else if val == nil {
			data.SyncEnabled = types.BoolNull()
		}
	}
	if val, ok := m["preventSearch"]; ok {
		if v, ok := val.(bool); ok {
			data.PreventSearch = types.BoolValue(v)
		} else if val == nil {
			data.PreventSearch = types.BoolNull()
		}
	}
	if val, ok := m["tagRequests"]; ok {
		if v, ok := val.(bool); ok {
			data.TagRequestsWithUser = types.BoolValue(v)
		} else if val == nil {
			data.TagRequestsWithUser = types.BoolNull()
		}
	}
	if data.APIKey.IsNull() || data.APIKey.IsUnknown() {
		if val, ok := m["apiKey"]; ok {
			if v, ok := val.(string); ok {
				data.APIKey = types.StringValue(v)
			}
		}
	}

	// tags is []float64 in JSON numbers
	if raw, ok := m["tags"]; ok {
		if raw == nil {
			data.Tags = types.ListNull(types.Int64Type)
		} else {
			var ids []int64
			if arr, ok := raw.([]any); ok {
				for _, el := range arr {
					if f, ok := el.(float64); ok {
						ids = append(ids, int64(f))
					}
				}
			}
			if ids == nil {
				ids = []int64{}
			}
			listVal, diags := types.ListValueFrom(ctx, types.Int64Type, ids)
			if diags.HasError() {
				return fmt.Errorf("build tags list: %s", diags[0].Detail())
			}
			data.Tags = listVal
		}
	}

	return nil
}

func (r *RadarrServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data RadarrServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/radarr", "", nil)
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
	if err := readRadarrStateFromJSON(ctx, item, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	// Preserve fields not returned by the API
	var state RadarrServerModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.APIKey = state.APIKey
	data.URL = state.URL
	data.ExtraPayloadJSON = state.ExtraPayloadJSON
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RadarrServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state RadarrServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var data RadarrServerModel
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
	path := fmt.Sprintf("/api/v1/settings/radarr/%d", data.ServerID.ValueInt64())
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
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RadarrServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data RadarrServerModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	path := fmt.Sprintf("/api/v1/settings/radarr/%d", data.ServerID.ValueInt64())
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

func (r *RadarrServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	id, err := requireInt64ID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Import Failed", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("server_id"), id)...)
}
