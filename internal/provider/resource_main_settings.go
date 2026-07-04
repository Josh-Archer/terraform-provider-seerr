package provider

import (
	"context"
	"encoding/json"
	"fmt"

	boolvalidator "github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	stringvalidator "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MainSettingsResource{}
var _ resource.ResourceWithImportState = &MainSettingsResource{}

type MainSettingsResource struct {
	client *APIClient
}

type MainSettingsModel struct {
	ID                    types.String `tfsdk:"id"`
	AppTitle              types.String `tfsdk:"app_title"`
	ApplicationURL        types.String `tfsdk:"application_url"`
	TrustProxy            types.Bool   `tfsdk:"trust_proxy"`
	CSRFProtection        types.Bool   `tfsdk:"csrf_protection"`
	ImageProxy            types.Bool   `tfsdk:"image_proxy"`
	CacheImages           types.Bool   `tfsdk:"cache_images"`
	Locale                types.String `tfsdk:"locale"`
	DiscoverRegion        types.String `tfsdk:"discover_region"`
	StreamingRegion       types.String `tfsdk:"streaming_region"`
	Region                types.String `tfsdk:"region"`
	OriginalLanguage      types.String `tfsdk:"original_language"`
	HideAvailable         types.Bool   `tfsdk:"hide_available"`
	PartialRequests       types.Bool   `tfsdk:"partial_requests"`
	LocalLogin            types.Bool   `tfsdk:"local_login"`
	MediaServerLogin      types.Bool   `tfsdk:"media_server_login"`
	NewPlexLogin          types.Bool   `tfsdk:"new_plex_login"`
	PlexLogin             types.Bool   `tfsdk:"plex_login"`
	MovieRequestsEnabled  types.Bool   `tfsdk:"movie_requests_enabled"`
	SeriesRequestsEnabled types.Bool   `tfsdk:"series_requests_enabled"`
	EnableReportAnIssue   types.Bool   `tfsdk:"enable_report_an_issue"`
	MovieRequestLimit     types.Int64  `tfsdk:"movie_request_limit"`
	SeriesRequestLimit    types.Int64  `tfsdk:"series_request_limit"`
	ResponseJSON          types.String `tfsdk:"response_json"`
	StatusCode            types.Int64  `tfsdk:"status_code"`
}

func NewMainSettingsResource() resource.Resource { return &MainSettingsResource{} }

func (r *MainSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_main_settings"
}

func (r *MainSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr main settings via /api/v1/settings/main.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_title": schema.StringAttribute{
				MarkdownDescription: "The application title.",
				Optional:            true,
				Computed:            true,
			},
			"application_url": schema.StringAttribute{
				MarkdownDescription: "The application URL.",
				Optional:            true,
				Computed:            true,
			},
			"trust_proxy": schema.BoolAttribute{
				MarkdownDescription: "Whether to trust the proxy.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "`trust_proxy` moved to `seerr_network_settings` in Seerr v3.",
			},
			"csrf_protection": schema.BoolAttribute{
				MarkdownDescription: "Whether CSRF protection is enabled.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "`csrf_protection` moved to `seerr_network_settings` in Seerr v3.",
			},
			"image_proxy": schema.BoolAttribute{
				MarkdownDescription: "Whether the image proxy is enabled.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Use `cache_images`; Seerr v3 exposes this setting as `cacheImages`.",
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.MatchRoot("cache_images")),
				},
			},
			"cache_images": schema.BoolAttribute{
				MarkdownDescription: "Whether Seerr caches proxied images.",
				Optional:            true,
				Computed:            true,
			},
			"locale": schema.StringAttribute{
				MarkdownDescription: "The application locale.",
				Optional:            true,
				Computed:            true,
			},
			"discover_region": schema.StringAttribute{
				MarkdownDescription: "The Discover region used by Seerr.",
				Optional:            true,
				Computed:            true,
			},
			"streaming_region": schema.StringAttribute{
				MarkdownDescription: "The streaming region used by Seerr.",
				Optional:            true,
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The application region.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Use `streaming_region` and `discover_region`; Seerr v3 split the legacy region field.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.MatchRoot("streaming_region")),
					stringvalidator.ConflictsWith(path.MatchRoot("discover_region")),
				},
			},
			"original_language": schema.StringAttribute{
				MarkdownDescription: "The original language.",
				Optional:            true,
				Computed:            true,
			},
			"hide_available": schema.BoolAttribute{
				MarkdownDescription: "Whether to hide available media.",
				Optional:            true,
				Computed:            true,
			},
			"partial_requests": schema.BoolAttribute{
				MarkdownDescription: "Whether partial requests are allowed.",
				Optional:            true,
				Computed:            true,
			},
			"local_login": schema.BoolAttribute{
				MarkdownDescription: "Whether local login is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"media_server_login": schema.BoolAttribute{
				MarkdownDescription: "Whether media-server login is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"new_plex_login": schema.BoolAttribute{
				MarkdownDescription: "Whether the new Plex login is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"plex_login": schema.BoolAttribute{
				MarkdownDescription: "Whether Plex login is enabled.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Use `media_server_login`; Seerr v3 generalized Plex login to media-server login.",
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.MatchRoot("media_server_login")),
				},
			},
			"movie_requests_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether movie requests are enabled.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Seerr v3 removed this main setting; use `partial_requests` and user quota settings instead.",
			},
			"series_requests_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether series requests are enabled.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Seerr v3 removed this main setting; use `partial_requests` and user quota settings instead.",
			},
			"enable_report_an_issue": schema.BoolAttribute{
				MarkdownDescription: "Whether the 'Report an Issue' feature is enabled.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Seerr v3 no longer exposes this value from `/api/v1/settings/main`.",
			},
			"movie_request_limit": schema.Int64Attribute{
				MarkdownDescription: "The movie request limit.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Seerr v3 moved request limits into user quota settings.",
			},
			"series_request_limit": schema.Int64Attribute{
				MarkdownDescription: "The series request limit.",
				Optional:            true,
				Computed:            true,
				DeprecationMessage:  "Seerr v3 moved request limits into user quota settings.",
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

func (r *MainSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MainSettingsResource) buildPayload(data *MainSettingsModel) map[string]any {
	payload := make(map[string]any)
	if !data.AppTitle.IsNull() && !data.AppTitle.IsUnknown() {
		payload["applicationTitle"] = data.AppTitle.ValueString()
	}
	if !data.ApplicationURL.IsNull() && !data.ApplicationURL.IsUnknown() {
		payload["applicationUrl"] = data.ApplicationURL.ValueString()
	}
	if !data.CacheImages.IsNull() && !data.CacheImages.IsUnknown() {
		payload["cacheImages"] = data.CacheImages.ValueBool()
	} else if !data.ImageProxy.IsNull() && !data.ImageProxy.IsUnknown() {
		payload["cacheImages"] = data.ImageProxy.ValueBool()
	}
	if !data.Locale.IsNull() && !data.Locale.IsUnknown() {
		payload["locale"] = data.Locale.ValueString()
	}
	if !data.DiscoverRegion.IsNull() && !data.DiscoverRegion.IsUnknown() {
		payload["discoverRegion"] = data.DiscoverRegion.ValueString()
	}
	if !data.StreamingRegion.IsNull() && !data.StreamingRegion.IsUnknown() {
		payload["streamingRegion"] = data.StreamingRegion.ValueString()
	} else if !data.Region.IsNull() && !data.Region.IsUnknown() {
		payload["streamingRegion"] = data.Region.ValueString()
	}
	if !data.OriginalLanguage.IsNull() && !data.OriginalLanguage.IsUnknown() {
		payload["originalLanguage"] = data.OriginalLanguage.ValueString()
	}
	if !data.HideAvailable.IsNull() && !data.HideAvailable.IsUnknown() {
		payload["hideAvailable"] = data.HideAvailable.ValueBool()
	}
	if !data.PartialRequests.IsNull() && !data.PartialRequests.IsUnknown() {
		payload["partialRequestsEnabled"] = data.PartialRequests.ValueBool()
	}
	if !data.LocalLogin.IsNull() && !data.LocalLogin.IsUnknown() {
		payload["localLogin"] = data.LocalLogin.ValueBool()
	}
	if !data.MediaServerLogin.IsNull() && !data.MediaServerLogin.IsUnknown() {
		payload["mediaServerLogin"] = data.MediaServerLogin.ValueBool()
	} else if !data.PlexLogin.IsNull() && !data.PlexLogin.IsUnknown() {
		payload["mediaServerLogin"] = data.PlexLogin.ValueBool()
	}
	if !data.NewPlexLogin.IsNull() && !data.NewPlexLogin.IsUnknown() {
		payload["newPlexLogin"] = data.NewPlexLogin.ValueBool()
	}

	return payload
}

type mainSettingsDecodedValues struct {
	AppTitle              types.String
	ApplicationURL        types.String
	TrustProxy            types.Bool
	CSRFProtection        types.Bool
	ImageProxy            types.Bool
	CacheImages           types.Bool
	Locale                types.String
	DiscoverRegion        types.String
	StreamingRegion       types.String
	Region                types.String
	OriginalLanguage      types.String
	HideAvailable         types.Bool
	PartialRequests       types.Bool
	LocalLogin            types.Bool
	MediaServerLogin      types.Bool
	NewPlexLogin          types.Bool
	PlexLogin             types.Bool
	MovieRequestsEnabled  types.Bool
	SeriesRequestsEnabled types.Bool
	EnableReportAnIssue   types.Bool
	MovieRequestLimit     types.Int64
	SeriesRequestLimit    types.Int64
}

func decodeMainSettings(decoded map[string]any) mainSettingsDecodedValues {
	values := mainSettingsDecodedValues{
		AppTitle:              types.StringNull(),
		ApplicationURL:        types.StringNull(),
		TrustProxy:            types.BoolNull(),
		CSRFProtection:        types.BoolNull(),
		ImageProxy:            types.BoolNull(),
		CacheImages:           types.BoolNull(),
		Locale:                types.StringNull(),
		DiscoverRegion:        types.StringNull(),
		StreamingRegion:       types.StringNull(),
		Region:                types.StringNull(),
		OriginalLanguage:      types.StringNull(),
		HideAvailable:         types.BoolNull(),
		PartialRequests:       types.BoolNull(),
		LocalLogin:            types.BoolNull(),
		MediaServerLogin:      types.BoolNull(),
		NewPlexLogin:          types.BoolNull(),
		PlexLogin:             types.BoolNull(),
		MovieRequestsEnabled:  types.BoolNull(),
		SeriesRequestsEnabled: types.BoolNull(),
		EnableReportAnIssue:   types.BoolNull(),
		MovieRequestLimit:     types.Int64Null(),
		SeriesRequestLimit:    types.Int64Null(),
	}

	if v, ok := stringValueFromAny(decoded["applicationTitle"]); ok {
		values.AppTitle = types.StringValue(v)
	} else if v, ok := stringValueFromAny(decoded["appTitle"]); ok {
		values.AppTitle = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["applicationUrl"]); ok {
		values.ApplicationURL = types.StringValue(v)
	}
	if v, ok := boolValueFromAny(decoded["trustProxy"]); ok {
		values.TrustProxy = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["csrfProtection"]); ok {
		values.CSRFProtection = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["cacheImages"]); ok {
		values.CacheImages = types.BoolValue(v)
		values.ImageProxy = types.BoolValue(v)
	} else if v, ok := boolValueFromAny(decoded["imageProxy"]); ok {
		values.ImageProxy = types.BoolValue(v)
	}
	if v, ok := stringValueFromAny(decoded["locale"]); ok {
		values.Locale = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["discoverRegion"]); ok {
		values.DiscoverRegion = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["streamingRegion"]); ok {
		values.StreamingRegion = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["streamingRegion"]); ok && v != "" {
		values.Region = types.StringValue(v)
	} else if v, ok := stringValueFromAny(decoded["discoverRegion"]); ok {
		values.Region = types.StringValue(v)
	} else if v, ok := stringValueFromAny(decoded["region"]); ok {
		values.Region = types.StringValue(v)
	}
	if v, ok := stringValueFromAny(decoded["originalLanguage"]); ok {
		values.OriginalLanguage = types.StringValue(v)
	}
	if v, ok := boolValueFromAny(decoded["hideAvailable"]); ok {
		values.HideAvailable = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["partialRequestsEnabled"]); ok {
		values.PartialRequests = types.BoolValue(v)
	} else if v, ok := boolValueFromAny(decoded["partialRequests"]); ok {
		values.PartialRequests = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["localLogin"]); ok {
		values.LocalLogin = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["mediaServerLogin"]); ok {
		values.MediaServerLogin = types.BoolValue(v)
		values.PlexLogin = types.BoolValue(v)
	} else if v, ok := boolValueFromAny(decoded["plexLogin"]); ok {
		values.PlexLogin = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["newPlexLogin"]); ok {
		values.NewPlexLogin = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["movieRequestsEnabled"]); ok {
		values.MovieRequestsEnabled = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["seriesRequestsEnabled"]); ok {
		values.SeriesRequestsEnabled = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["enableReportAnIssue"]); ok {
		values.EnableReportAnIssue = types.BoolValue(v)
	}
	if v, ok := int64ValueFromAny(decoded["movieRequestLimit"]); ok {
		values.MovieRequestLimit = types.Int64Value(v)
	}
	if v, ok := int64ValueFromAny(decoded["seriesRequestLimit"]); ok {
		values.SeriesRequestLimit = types.Int64Value(v)
	}

	return values
}

func (r *MainSettingsResource) applyDecodedSettings(data *MainSettingsModel, decoded map[string]any) {
	values := decodeMainSettings(decoded)
	data.AppTitle = values.AppTitle
	data.ApplicationURL = values.ApplicationURL
	data.TrustProxy = values.TrustProxy
	data.CSRFProtection = values.CSRFProtection
	data.ImageProxy = values.ImageProxy
	data.CacheImages = values.CacheImages
	data.Locale = values.Locale
	data.DiscoverRegion = values.DiscoverRegion
	data.StreamingRegion = values.StreamingRegion
	data.Region = values.Region
	data.OriginalLanguage = values.OriginalLanguage
	data.HideAvailable = values.HideAvailable
	data.PartialRequests = values.PartialRequests
	data.LocalLogin = values.LocalLogin
	data.MediaServerLogin = values.MediaServerLogin
	data.NewPlexLogin = values.NewPlexLogin
	data.PlexLogin = values.PlexLogin
	data.MovieRequestsEnabled = values.MovieRequestsEnabled
	data.SeriesRequestsEnabled = values.SeriesRequestsEnabled
	data.EnableReportAnIssue = values.EnableReportAnIssue
	data.MovieRequestLimit = values.MovieRequestLimit
	data.SeriesRequestLimit = values.SeriesRequestLimit
}

func (r *MainSettingsResource) refreshState(ctx context.Context, data *MainSettingsModel) error {
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/main", "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		return fmt.Errorf("failed to decode response: %s", err)
	}

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))
	r.applyDecodedSettings(data, decoded)
	return nil
}

func (r *MainSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MainSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.buildPayload(&data))
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/main", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Create") {
		return
	}

	data.ID = types.StringValue("main")
	if err := r.refreshState(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MainSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MainSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.refreshState(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.ID = types.StringValue("main")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MainSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MainSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := json.Marshal(r.buildPayload(&data))
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/main", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Update") {
		return
	}

	data.ID = types.StringValue("main")
	if err := r.refreshState(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MainSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// There is no concept of "deleting" main settings in Seerr; it is a singleton.
	// This method only removes the resource from Terraform state.
	// The settings remain as-is on the Seerr instance.
}

func (r *MainSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
