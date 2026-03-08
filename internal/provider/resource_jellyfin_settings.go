package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &JellyfinSettingsResource{}
var _ resource.ResourceWithImportState = &JellyfinSettingsResource{}

type JellyfinSettingsResource struct {
	client *APIClient
}

type JellyfinSettingsModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	IP                        types.String `tfsdk:"ip"`
	Port                      types.Int64  `tfsdk:"port"`
	UseSSL                    types.Bool   `tfsdk:"use_ssl"`
	URLBase                   types.String `tfsdk:"url_base"`
	ExternalHostname          types.String `tfsdk:"external_hostname"`
	JellyfinForgotPasswordURL types.String `tfsdk:"jellyfin_forgot_password_url"`
	ServerID                  types.String `tfsdk:"server_id"`
	APIKey                    types.String `tfsdk:"api_key"`
	ResponseJSON              types.String `tfsdk:"response_json"`
	StatusCode                types.Int64  `tfsdk:"status_code"`
}

func NewJellyfinSettingsResource() resource.Resource { return &JellyfinSettingsResource{} }

func (r *JellyfinSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_jellyfin_settings"
}

func (r *JellyfinSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr Jellyfin settings via /api/v1/settings/jellyfin.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Jellyfin server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Jellyfin"),
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address or hostname of the Jellyfin server.",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port of the Jellyfin server.",
				Required:            true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Whether to use SSL for the connection.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"url_base": schema.StringAttribute{
				MarkdownDescription: "The base URL for the Jellyfin server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"external_hostname": schema.StringAttribute{
				MarkdownDescription: "The external hostname for the Jellyfin server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"jellyfin_forgot_password_url": schema.StringAttribute{
				MarkdownDescription: "The URL for forgotten passwords on the Jellyfin server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for the Jellyfin server.",
				Required:            true,
				Sensitive:           true,
			},
			"server_id": schema.StringAttribute{
				MarkdownDescription: "The unique server ID of the connected Jellyfin server.",
				Computed:            true,
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

func (r *JellyfinSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *JellyfinSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data JellyfinSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"ip":     data.IP.ValueString(),
		"port":   data.Port.ValueInt64(),
		"apiKey": data.APIKey.ValueString(),
	}
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		payload["name"] = data.Name.ValueString()
	}
	if !data.UseSSL.IsNull() && !data.UseSSL.IsUnknown() {
		payload["useSsl"] = data.UseSSL.ValueBool()
	}
	if !data.URLBase.IsNull() && !data.URLBase.IsUnknown() {
		payload["urlBase"] = data.URLBase.ValueString()
	}
	if !data.ExternalHostname.IsNull() && !data.ExternalHostname.IsUnknown() {
		payload["externalHostname"] = data.ExternalHostname.ValueString()
	}
	if !data.JellyfinForgotPasswordURL.IsNull() && !data.JellyfinForgotPasswordURL.IsUnknown() {
		payload["jellyfinForgotPasswordUrl"] = data.JellyfinForgotPasswordURL.ValueString()
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/jellyfin", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue("jellyfin")
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	// Refresh state from response
	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}
	if v, ok := decoded["name"].(string); ok {
		data.Name = types.StringValue(v)
	}
	if v, ok := decoded["ip"].(string); ok {
		data.IP = types.StringValue(v)
	}
	if v, ok := decoded["port"].(float64); ok {
		data.Port = types.Int64Value(int64(v))
	}
	if v, ok := decoded["useSsl"].(bool); ok {
		data.UseSSL = types.BoolValue(v)
	}
	if v, ok := decoded["urlBase"].(string); ok {
		data.URLBase = types.StringValue(v)
	}
	if v, ok := decoded["externalHostname"].(string); ok {
		data.ExternalHostname = types.StringValue(v)
	}
	if v, ok := decoded["jellyfinForgotPasswordUrl"].(string); ok {
		data.JellyfinForgotPasswordURL = types.StringValue(v)
	}
	if v, ok := decoded["serverId"].(string); ok {
		data.ServerID = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JellyfinSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data JellyfinSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/jellyfin", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	if v, ok := decoded["name"].(string); ok {
		data.Name = types.StringValue(v)
	}
	if v, ok := decoded["ip"].(string); ok {
		data.IP = types.StringValue(v)
	}
	if v, ok := decoded["port"].(float64); ok {
		data.Port = types.Int64Value(int64(v))
	}
	if v, ok := decoded["useSsl"].(bool); ok {
		data.UseSSL = types.BoolValue(v)
	}
	if v, ok := decoded["urlBase"].(string); ok {
		data.URLBase = types.StringValue(v)
	}
	if v, ok := decoded["externalHostname"].(string); ok {
		data.ExternalHostname = types.StringValue(v)
	}
	if v, ok := decoded["jellyfinForgotPasswordUrl"].(string); ok {
		data.JellyfinForgotPasswordURL = types.StringValue(v)
	}
	// the API key might not always be returned in the read response but usually is in Jellyseerr API
	if v, ok := decoded["apiKey"].(string); ok && v != "" {
		data.APIKey = types.StringValue(v)
	}
	if v, ok := decoded["serverId"].(string); ok {
		data.ServerID = types.StringValue(v)
	}

	data.ID = types.StringValue("jellyfin")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JellyfinSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data JellyfinSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"ip":     data.IP.ValueString(),
		"port":   data.Port.ValueInt64(),
		"apiKey": data.APIKey.ValueString(),
	}
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		payload["name"] = data.Name.ValueString()
	}
	if !data.UseSSL.IsNull() && !data.UseSSL.IsUnknown() {
		payload["useSsl"] = data.UseSSL.ValueBool()
	}
	if !data.URLBase.IsNull() && !data.URLBase.IsUnknown() {
		payload["urlBase"] = data.URLBase.ValueString()
	}
	if !data.ExternalHostname.IsNull() && !data.ExternalHostname.IsUnknown() {
		payload["externalHostname"] = data.ExternalHostname.ValueString()
	}
	if !data.JellyfinForgotPasswordURL.IsNull() && !data.JellyfinForgotPasswordURL.IsUnknown() {
		payload["jellyfinForgotPasswordUrl"] = data.JellyfinForgotPasswordURL.ValueString()
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/jellyfin", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue("jellyfin")
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	// Refresh state from response
	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}
	if v, ok := decoded["name"].(string); ok {
		data.Name = types.StringValue(v)
	}
	if v, ok := decoded["ip"].(string); ok {
		data.IP = types.StringValue(v)
	}
	if v, ok := decoded["port"].(float64); ok {
		data.Port = types.Int64Value(int64(v))
	}
	if v, ok := decoded["useSsl"].(bool); ok {
		data.UseSSL = types.BoolValue(v)
	}
	if v, ok := decoded["urlBase"].(string); ok {
		data.URLBase = types.StringValue(v)
	}
	if v, ok := decoded["externalHostname"].(string); ok {
		data.ExternalHostname = types.StringValue(v)
	}
	if v, ok := decoded["jellyfinForgotPasswordUrl"].(string); ok {
		data.JellyfinForgotPasswordURL = types.StringValue(v)
	}
	if v, ok := decoded["serverId"].(string); ok {
		data.ServerID = types.StringValue(v)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *JellyfinSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No DELETE route for Jellyfin settings.
}

func (r *JellyfinSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
