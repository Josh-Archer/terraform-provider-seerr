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

var _ resource.Resource = &TautulliSettingsResource{}
var _ resource.ResourceWithImportState = &TautulliSettingsResource{}

type TautulliSettingsResource struct {
	client *APIClient
}

type TautulliSettingsModel struct {
	ID           types.String `tfsdk:"id"`
	Hostname     types.String `tfsdk:"hostname"`
	Port         types.Int64  `tfsdk:"port"`
	UseSSL       types.Bool   `tfsdk:"use_ssl"`
	URLBase      types.String `tfsdk:"url_base"`
	APIKey       types.String `tfsdk:"api_key"`
	ExternalURL  types.String `tfsdk:"external_url"`
	ResponseJSON types.String `tfsdk:"response_json"`
	StatusCode   types.Int64  `tfsdk:"status_code"`
}

func NewTautulliSettingsResource() resource.Resource { return &TautulliSettingsResource{} }

func (r *TautulliSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tautulli_settings"
}

func (r *TautulliSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manage Seerr Tautulli settings via /api/v1/settings/tautulli.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "The hostname of the Tautulli server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port of the Tautulli server.",
				Optional:            true,
				Computed:            true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Whether to use SSL for the connection.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"url_base": schema.StringAttribute{
				MarkdownDescription: "The base URL for the Tautulli server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for the Tautulli server.",
				Optional:            true,
				Sensitive:           true,
			},
			"external_url": schema.StringAttribute{
				MarkdownDescription: "The external URL for the Tautulli server.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
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

func (r *TautulliSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *TautulliSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TautulliSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := make(map[string]any)
	if !data.Hostname.IsNull() && !data.Hostname.IsUnknown() {
		payload["hostname"] = data.Hostname.ValueString()
	}
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		payload["port"] = data.Port.ValueInt64()
	}
	if !data.UseSSL.IsNull() && !data.UseSSL.IsUnknown() {
		payload["useSsl"] = data.UseSSL.ValueBool()
	}
	if !data.URLBase.IsNull() && !data.URLBase.IsUnknown() {
		payload["urlBase"] = data.URLBase.ValueString()
	}
	if !data.APIKey.IsNull() && !data.APIKey.IsUnknown() {
		payload["apiKey"] = data.APIKey.ValueString()
	}
	if !data.ExternalURL.IsNull() && !data.ExternalURL.IsUnknown() {
		payload["externalUrl"] = data.ExternalURL.ValueString()
	}

	body, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("failed to marshal payload: %s", err))
		return
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/tautulli", string(body), nil)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Create Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	data.ID = types.StringValue("tautulli")
	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	// Refresh state from response
	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err == nil {
		if v, ok := decoded["hostname"].(string); ok {
			data.Hostname = types.StringValue(v)
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
		if v, ok := decoded["apiKey"].(string); ok && v != "" {
			data.APIKey = types.StringValue(v)
		}
		if v, ok := decoded["externalUrl"].(string); ok {
			data.ExternalURL = types.StringValue(v)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TautulliSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TautulliSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/tautulli", "", nil)
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
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.StatusCode = types.Int64Value(int64(res.StatusCode))
	data.ResponseJSON = types.StringValue(string(res.Body))

	if v, ok := decoded["hostname"].(string); ok {
		data.Hostname = types.StringValue(v)
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
	if v, ok := decoded["apiKey"].(string); ok && v != "" {
		data.APIKey = types.StringValue(v)
	}
	if v, ok := decoded["externalUrl"].(string); ok {
		data.ExternalURL = types.StringValue(v)
	}

	data.ID = types.StringValue("tautulli")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TautulliSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	r.Create(ctx, resource.CreateRequest{Plan: req.Plan}, &resource.CreateResponse{State: resp.State, Diagnostics: resp.Diagnostics})
}

func (r *TautulliSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *TautulliSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
