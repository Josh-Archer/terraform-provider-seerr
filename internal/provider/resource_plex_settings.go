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

var _ resource.Resource = &PlexSettingsResource{}
var _ resource.ResourceWithImportState = &PlexSettingsResource{}

type PlexSettingsResource struct {
	client *APIClient
}

type PlexSettingsModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	IP           types.String `tfsdk:"ip"`
	Port         types.Int64  `tfsdk:"port"`
	UseSSL       types.Bool   `tfsdk:"use_ssl"`
	ResponseJSON types.String `tfsdk:"response_json"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
}

func NewPlexSettingsResource() resource.Resource { return &PlexSettingsResource{} }

func (r *PlexSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plex_settings"
}

func (r *PlexSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr Plex settings via /api/v1/settings/plex.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the Plex server.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ip": schema.StringAttribute{
				MarkdownDescription: "The IP address or hostname of the Plex server.",
				Required:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port of the Plex server.",
				Required:            true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Whether to use SSL for the connection.",
				Optional:            true,
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

func (r *PlexSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PlexSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PlexSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"ip":   data.IP.ValueString(),
		"port": data.Port.ValueInt64(),
	}
	if !data.UseSSL.IsNull() && !data.UseSSL.IsUnknown() {
		payload["useSsl"] = data.UseSSL.ValueBool()
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/plex", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue("plex")
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlexSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PlexSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/plex", "", nil)
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

	data.ID = types.StringValue("plex")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlexSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PlexSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{
		"ip":   data.IP.ValueString(),
		"port": data.Port.ValueInt64(),
	}
	if !data.UseSSL.IsNull() && !data.UseSSL.IsUnknown() {
		payload["useSsl"] = data.UseSSL.ValueBool()
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/plex", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Update Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue("plex")
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

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlexSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// There is no concept of "deleting" Plex settings in Seerr; it is a singleton.
	// This method only removes the resource from Terraform state.
	// The settings remain as-is on the Seerr instance.
}

func (r *PlexSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
