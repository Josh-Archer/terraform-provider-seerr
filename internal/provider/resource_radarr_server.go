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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	ActiveProfileID     types.Int64  `tfsdk:"active_profile_id"`
	ActiveDirectory     types.String `tfsdk:"active_directory"`
	Is4K                types.Bool   `tfsdk:"is_4k"`
	MinimumAvailability types.String `tfsdk:"minimum_availability"`
	Tags                types.List   `tfsdk:"tags"`
	IsDefault           types.Bool   `tfsdk:"is_default"`
	SyncEnabled         types.Bool   `tfsdk:"sync_enabled"`
	PreventSearch       types.Bool   `tfsdk:"prevent_search"`
	TagRequests         types.Bool   `tfsdk:"tag_requests"`
	ExtraPayloadJSON    types.String `tfsdk:"extra_payload_json"`
	ResponseJSON        types.String `tfsdk:"response_json"`
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
			"server_id": schema.Int64Attribute{Computed: true},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("Radarr"),
			},
			"url": schema.StringAttribute{Optional: true},
			"hostname": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("radarr-service"),
			},
			"port": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
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
			"active_profile_id": schema.Int64Attribute{Required: true},
			"active_directory":  schema.StringAttribute{Required: true},
			"is_4k": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"minimum_availability": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("announced"),
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.Int64Type,
			},
			"is_default": schema.BoolAttribute{
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
			"tag_requests": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"extra_payload_json": schema.StringAttribute{Optional: true},
			"response_json":      schema.StringAttribute{Computed: true},
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

func (r *RadarrServerResource) payload(ctx context.Context, data RadarrServerModel) (string, error) {
	parseURLIntoModel(&data)
	tags, err := listInt64(ctx, data.Tags)
	if err != nil {
		return "", err
	}

	base := map[string]any{
		"name":                data.Name.ValueString(),
		"hostname":            data.Hostname.ValueString(),
		"port":                data.Port.ValueInt64(),
		"apiKey":              data.APIKey.ValueString(),
		"useSsl":              data.UseSSL.ValueBool(),
		"baseUrl":             data.BaseURL.ValueString(),
		"activeProfileId":     data.ActiveProfileID.ValueInt64(),
		"activeDirectory":     data.ActiveDirectory.ValueString(),
		"is4k":                data.Is4K.ValueBool(),
		"minimumAvailability": data.MinimumAvailability.ValueString(),
		"tags":                tags,
		"isDefault":           data.IsDefault.ValueBool(),
		"syncEnabled":         data.SyncEnabled.ValueBool(),
		"preventSearch":       data.PreventSearch.ValueBool(),
		"tagRequests":         data.TagRequests.ValueBool(),
	}
	merged, err := mergeJSON(base, data.ExtraPayloadJSON.ValueString())
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(merged)
	return string(b), err
}

func (r *RadarrServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data RadarrServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := r.payload(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}
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
	data.ResponseJSON = types.StringValue(string(res.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
	data.ResponseJSON = types.StringValue(string(item))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *RadarrServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data RadarrServerModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := r.payload(ctx, data)
	if err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}
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
	data.ResponseJSON = types.StringValue(string(res.Body))
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
