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
	Locale                types.String `tfsdk:"locale"`
	Region                types.String `tfsdk:"region"`
	OriginalLanguage      types.String `tfsdk:"original_language"`
	HideAvailable         types.Bool   `tfsdk:"hide_available"`
	PartialRequests       types.Bool   `tfsdk:"partial_requests"`
	LocalLogin            types.Bool   `tfsdk:"local_login"`
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
			},
			"csrf_protection": schema.BoolAttribute{
				MarkdownDescription: "Whether CSRF protection is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"image_proxy": schema.BoolAttribute{
				MarkdownDescription: "Whether the image proxy is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"locale": schema.StringAttribute{
				MarkdownDescription: "The application locale.",
				Optional:            true,
				Computed:            true,
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "The application region.",
				Optional:            true,
				Computed:            true,
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
			"new_plex_login": schema.BoolAttribute{
				MarkdownDescription: "Whether the new Plex login is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"plex_login": schema.BoolAttribute{
				MarkdownDescription: "Whether Plex login is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"movie_requests_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether movie requests are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"series_requests_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether series requests are enabled.",
				Optional:            true,
				Computed:            true,
			},
			"enable_report_an_issue": schema.BoolAttribute{
				MarkdownDescription: "Whether the 'Report an Issue' feature is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"movie_request_limit": schema.Int64Attribute{
				MarkdownDescription: "The movie request limit.",
				Optional:            true,
				Computed:            true,
			},
			"series_request_limit": schema.Int64Attribute{
				MarkdownDescription: "The series request limit.",
				Optional:            true,
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
		payload["appTitle"] = data.AppTitle.ValueString()
	}
	if !data.ApplicationURL.IsNull() && !data.ApplicationURL.IsUnknown() {
		payload["applicationUrl"] = data.ApplicationURL.ValueString()
	}
	if !data.TrustProxy.IsNull() && !data.TrustProxy.IsUnknown() {
		payload["trustProxy"] = data.TrustProxy.ValueBool()
	}
	if !data.CSRFProtection.IsNull() && !data.CSRFProtection.IsUnknown() {
		payload["csrfProtection"] = data.CSRFProtection.ValueBool()
	}
	if !data.ImageProxy.IsNull() && !data.ImageProxy.IsUnknown() {
		payload["imageProxy"] = data.ImageProxy.ValueBool()
	}
	if !data.Locale.IsNull() && !data.Locale.IsUnknown() {
		payload["locale"] = data.Locale.ValueString()
	}
	if !data.Region.IsNull() && !data.Region.IsUnknown() {
		payload["region"] = data.Region.ValueString()
	}
	if !data.OriginalLanguage.IsNull() && !data.OriginalLanguage.IsUnknown() {
		payload["originalLanguage"] = data.OriginalLanguage.ValueString()
	}
	if !data.HideAvailable.IsNull() && !data.HideAvailable.IsUnknown() {
		payload["hideAvailable"] = data.HideAvailable.ValueBool()
	}
	if !data.PartialRequests.IsNull() && !data.PartialRequests.IsUnknown() {
		payload["partialRequests"] = data.PartialRequests.ValueBool()
	}
	if !data.LocalLogin.IsNull() && !data.LocalLogin.IsUnknown() {
		payload["localLogin"] = data.LocalLogin.ValueBool()
	}
	if !data.NewPlexLogin.IsNull() && !data.NewPlexLogin.IsUnknown() {
		payload["newPlexLogin"] = data.NewPlexLogin.ValueBool()
	}
	if !data.PlexLogin.IsNull() && !data.PlexLogin.IsUnknown() {
		payload["plexLogin"] = data.PlexLogin.ValueBool()
	}
	if !data.MovieRequestsEnabled.IsNull() && !data.MovieRequestsEnabled.IsUnknown() {
		payload["movieRequestsEnabled"] = data.MovieRequestsEnabled.ValueBool()
	}
	if !data.SeriesRequestsEnabled.IsNull() && !data.SeriesRequestsEnabled.IsUnknown() {
		payload["seriesRequestsEnabled"] = data.SeriesRequestsEnabled.ValueBool()
	}
	if !data.EnableReportAnIssue.IsNull() && !data.EnableReportAnIssue.IsUnknown() {
		payload["enableReportAnIssue"] = data.EnableReportAnIssue.ValueBool()
	}
	if !data.MovieRequestLimit.IsNull() && !data.MovieRequestLimit.IsUnknown() {
		payload["movieRequestLimit"] = data.MovieRequestLimit.ValueInt64()
	}
	if !data.SeriesRequestLimit.IsNull() && !data.SeriesRequestLimit.IsUnknown() {
		payload["seriesRequestLimit"] = data.SeriesRequestLimit.ValueInt64()
	}

	return payload
}

func (r *MainSettingsResource) applyDecodedSettings(data *MainSettingsModel, decoded map[string]any) {
	if v, ok := decoded["appTitle"].(string); ok {
		data.AppTitle = types.StringValue(v)
	}
	if v, ok := decoded["applicationUrl"].(string); ok {
		data.ApplicationURL = types.StringValue(v)
	}
	if v, ok := decoded["trustProxy"].(bool); ok {
		data.TrustProxy = types.BoolValue(v)
	}
	if v, ok := decoded["csrfProtection"].(bool); ok {
		data.CSRFProtection = types.BoolValue(v)
	}
	if v, ok := decoded["imageProxy"].(bool); ok {
		data.ImageProxy = types.BoolValue(v)
	}
	if v, ok := decoded["locale"].(string); ok {
		data.Locale = types.StringValue(v)
	}
	if v, ok := decoded["region"].(string); ok {
		data.Region = types.StringValue(v)
	}
	if v, ok := decoded["originalLanguage"].(string); ok {
		data.OriginalLanguage = types.StringValue(v)
	}
	if v, ok := decoded["hideAvailable"].(bool); ok {
		data.HideAvailable = types.BoolValue(v)
	}
	if v, ok := decoded["partialRequests"].(bool); ok {
		data.PartialRequests = types.BoolValue(v)
	}
	if v, ok := decoded["localLogin"].(bool); ok {
		data.LocalLogin = types.BoolValue(v)
	}
	if v, ok := decoded["newPlexLogin"].(bool); ok {
		data.NewPlexLogin = types.BoolValue(v)
	}
	if v, ok := decoded["plexLogin"].(bool); ok {
		data.PlexLogin = types.BoolValue(v)
	}
	if v, ok := decoded["movieRequestsEnabled"].(bool); ok {
		data.MovieRequestsEnabled = types.BoolValue(v)
	}
	if v, ok := decoded["seriesRequestsEnabled"].(bool); ok {
		data.SeriesRequestsEnabled = types.BoolValue(v)
	}
	if v, ok := decoded["enableReportAnIssue"].(bool); ok {
		data.EnableReportAnIssue = types.BoolValue(v)
	}
	if v, ok := decoded["movieRequestLimit"].(float64); ok {
		data.MovieRequestLimit = types.Int64Value(int64(v))
	}
	if v, ok := decoded["seriesRequestLimit"].(float64); ok {
		data.SeriesRequestLimit = types.Int64Value(int64(v))
	}
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
