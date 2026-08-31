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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("hostname")),
					stringvalidator.ConflictsWith(path.MatchRoot("port")),
					stringvalidator.RegexMatches(urlRegex(), "must be a valid HTTP or HTTPS URL (e.g., http://localhost:8989)"),
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
			"active_anime_directory": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"anime_tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
			},
			"is_4k": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
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
			"enable_season_folders": schema.BoolAttribute{
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

	name := "Sonarr"
	if !data.Name.IsNull() && !data.Name.IsUnknown() && strings.TrimSpace(data.Name.ValueString()) != "" {
		name = strings.TrimSpace(data.Name.ValueString())
	}
	data.Name = types.StringValue(name)

	hostname := "sonarr-service"
	if !data.Hostname.IsNull() && !data.Hostname.IsUnknown() && strings.TrimSpace(data.Hostname.ValueString()) != "" {
		hostname = strings.TrimSpace(data.Hostname.ValueString())
	}
	data.Hostname = types.StringValue(hostname)

	port := int64(8989)
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

	isDefault := true
	if !data.IsDefault.IsNull() && !data.IsDefault.IsUnknown() {
		isDefault = data.IsDefault.ValueBool()
	}
	data.IsDefault = types.BoolValue(isDefault)

	enableSeasonFolders := true
	if !data.EnableSeasonFolders.IsNull() && !data.EnableSeasonFolders.IsUnknown() {
		enableSeasonFolders = data.EnableSeasonFolders.ValueBool()
	}
	data.EnableSeasonFolders = types.BoolValue(enableSeasonFolders)

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

	if profileName == "" && !data.QualityProfileID.IsNull() && !data.QualityProfileID.IsUnknown() {
		profileID := data.QualityProfileID.ValueInt64()
		// Try resolving via Seerr proxy test endpoint first
		if r.client != nil {
			testBody := map[string]any{
				"hostname": hostname,
				"port":     port,
				"apiKey":   data.APIKey.ValueString(),
				"useSsl":   useSSL,
				"baseUrl":  baseURL,
			}
			if testJSON, err := json.Marshal(testBody); err == nil {
				if testResp, err := r.client.Request(ctx, "POST", "/api/v1/settings/sonarr/test", string(testJSON), nil); err == nil && testResp.StatusCode >= 200 && testResp.StatusCode < 300 {
					var testResult struct {
						Profiles []struct {
							ID   int64  `json:"id"`
							Name string `json:"name"`
						} `json:"profiles"`
					}
					if err := json.Unmarshal(testResp.Body, &testResult); err == nil {
						for _, p := range testResult.Profiles {
							if p.ID == profileID {
								profileName = strings.TrimSpace(p.Name)
								break
							}
						}
					}
				}
			}
		}

		// If not resolved via Seerr proxy test, try direct Arr profile lookup fallback
		if profileName == "" {
			timeout := defaultRequestTimeout
			if r.client != nil {
				timeout = r.client.Timeout()
			}
			if profile, lookupErr := findArrProfile(
				ctx,
				data.URL.ValueString(),
				hostname,
				port,
				useSSL,
				baseURL,
				data.APIKey.ValueString(),
				timeout,
				&profileID,
				nil,
			); lookupErr == nil && profile != nil {
				profileName = profile.Name
			}
		}

		if profileName == "" {
			return data, "", fmt.Errorf("could not resolve quality_profile_name for profile id %d; please specify quality_profile_name explicitly", profileID)
		}
	}
	data.QualityProfileName = types.StringValue(profileName)

	base := map[string]any{
		"name":                 name,
		"hostname":             hostname,
		"port":                 port,
		"apiKey":               data.APIKey.ValueString(),
		"useSsl":               useSSL,
		"baseUrl":              baseURL,
		"activeProfileId":      data.QualityProfileID.ValueInt64(),
		"activeProfileName":    profileName,
		"activeDirectory":      data.ActiveDirectory.ValueString(),
		"activeAnimeDirectory": animeDir,
		"tags":                 tags,
		"animeTags":            animeTags,
		"is4k":                 is4K,
		"isDefault":            isDefault,
		"enableSeasonFolders":  enableSeasonFolders,
		"syncEnabled":          syncEnabled,
		"preventSearch":        preventSearch,
		"tagRequests":          tagRequests,
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
	if err := readSonarrStateFromJSON(ctx, res.Body, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("read state after create: %s", err))
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// readSonarrStateFromJSON populates all API-sourced fields of data from the
// JSON object representing a single Sonarr server entry returned by Overseerr.
// Fields that are user-supplied and never echoed by the API (url and
// extra_payload_json) are intentionally left untouched. api_key is preserved
// from state when it is already set, but data source reads can still populate
// it when the API returns a value.
func readSonarrStateFromJSON(ctx context.Context, item []byte, data *SonarrServerModel) error {
	var m map[string]any
	if err := json.Unmarshal(item, &m); err != nil {
		return fmt.Errorf("parse sonarr server response: %w", err)
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
	if val, ok := m["activeAnimeDirectory"]; ok {
		if v, ok := val.(string); ok {
			data.ActiveAnimeDirectory = types.StringValue(v)
		} else if val == nil {
			data.ActiveAnimeDirectory = types.StringNull()
		}
	}
	if val, ok := m["is4k"]; ok {
		if v, ok := val.(bool); ok {
			data.Is4K = types.BoolValue(v)
		} else if val == nil {
			data.Is4K = types.BoolNull()
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
	if val, ok := m["enableSeasonFolders"]; ok {
		if v, ok := val.(bool); ok {
			data.EnableSeasonFolders = types.BoolValue(v)
		} else if val == nil {
			data.EnableSeasonFolders = types.BoolNull()
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

	// tags
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

	// animeTags
	if raw, ok := m["animeTags"]; ok {
		if raw == nil {
			data.AnimeTags = types.ListNull(types.Int64Type)
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
				return fmt.Errorf("build animeTags list: %s", diags[0].Detail())
			}
			data.AnimeTags = listVal
		}
	}

	if data.Name.IsUnknown() {
		data.Name = types.StringNull()
	}
	if data.Hostname.IsUnknown() {
		data.Hostname = types.StringNull()
	}
	if data.Port.IsUnknown() {
		data.Port = types.Int64Null()
	}
	if data.UseSSL.IsUnknown() {
		data.UseSSL = types.BoolNull()
	}
	if data.BaseURL.IsUnknown() {
		data.BaseURL = types.StringNull()
	}
	if data.QualityProfileID.IsUnknown() {
		data.QualityProfileID = types.Int64Null()
	}
	if data.QualityProfileName.IsUnknown() {
		data.QualityProfileName = types.StringNull()
	}
	if data.ActiveDirectory.IsUnknown() {
		data.ActiveDirectory = types.StringNull()
	}
	if data.ActiveAnimeDirectory.IsUnknown() {
		data.ActiveAnimeDirectory = types.StringNull()
	}
	if data.Is4K.IsUnknown() {
		data.Is4K = types.BoolNull()
	}
	if data.IsDefault.IsUnknown() {
		data.IsDefault = types.BoolNull()
	}
	if data.EnableScan.IsUnknown() {
		data.EnableScan = types.BoolNull()
	}
	if data.EnableSeasonFolders.IsUnknown() {
		data.EnableSeasonFolders = types.BoolNull()
	}
	if data.SyncEnabled.IsUnknown() {
		data.SyncEnabled = types.BoolNull()
	}
	if data.PreventSearch.IsUnknown() {
		data.PreventSearch = types.BoolNull()
	}
	if data.TagRequestsWithUser.IsUnknown() {
		data.TagRequestsWithUser = types.BoolNull()
	}
	if data.APIKey.IsUnknown() {
		data.APIKey = types.StringNull()
	}
	if data.Tags.IsUnknown() {
		data.Tags = types.ListNull(types.Int64Type)
	}
	if data.AnimeTags.IsUnknown() {
		data.AnimeTags = types.ListNull(types.Int64Type)
	}
	if data.URL.IsUnknown() {
		data.URL = types.StringNull()
	}
	if data.ExtraPayloadJSON.IsUnknown() {
		data.ExtraPayloadJSON = types.StringNull()
	}

	return nil
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
	if res.StatusCode == 404 {
		resp.State.RemoveResource(ctx)
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
	if err := readSonarrStateFromJSON(ctx, item, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	// Preserve fields not returned by the API
	var state SonarrServerModel
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
	if err := readSonarrStateFromJSON(ctx, res.Body, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	data.ID = types.StringValue(fmt.Sprintf("%d", data.ServerID.ValueInt64()))
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
