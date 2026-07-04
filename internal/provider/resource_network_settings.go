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

var _ resource.Resource = &NetworkSettingsResource{}
var _ resource.ResourceWithImportState = &NetworkSettingsResource{}

type NetworkSettingsResource struct {
	client *APIClient
}

type NetworkProxyModel struct {
	Enabled              types.Bool   `tfsdk:"enabled"`
	Hostname             types.String `tfsdk:"hostname"`
	Port                 types.Int64  `tfsdk:"port"`
	UseSSL               types.Bool   `tfsdk:"use_ssl"`
	User                 types.String `tfsdk:"user"`
	Password             types.String `tfsdk:"password"`
	BypassFilter         types.String `tfsdk:"bypass_filter"`
	BypassLocalAddresses types.Bool   `tfsdk:"bypass_local_addresses"`
}

type NetworkDNSCacheModel struct {
	Enabled     types.Bool  `tfsdk:"enabled"`
	ForceMinTTL types.Int64 `tfsdk:"force_min_ttl"`
	ForceMaxTTL types.Int64 `tfsdk:"force_max_ttl"`
}

type NetworkSettingsModel struct {
	ID                  types.String          `tfsdk:"id"`
	CSRFProtection      types.Bool            `tfsdk:"csrf_protection"`
	ForceIPv4First      types.Bool            `tfsdk:"force_ipv4_first"`
	TrustProxy          types.Bool            `tfsdk:"trust_proxy"`
	APIRequestTimeoutMS types.Int64           `tfsdk:"api_request_timeout_ms"`
	Proxy               *NetworkProxyModel    `tfsdk:"proxy"`
	DNSCache            *NetworkDNSCacheModel `tfsdk:"dns_cache"`
}

func NewNetworkSettingsResource() resource.Resource { return &NetworkSettingsResource{} }

func (r *NetworkSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_settings"
}

func networkSettingsResourceSchema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manage Seerr network settings via `/api/v1/settings/network`.\n\n" +
			"`proxy` and `dns_cache` are opt-in managed blocks. Seerr may return default values for these nested settings even when they are not configured, but the provider leaves them absent from resource state unless the block is declared in configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"csrf_protection": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"force_ipv4_first": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"trust_proxy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
			"api_request_timeout_ms": schema.Int64Attribute{
				MarkdownDescription: "Maximum time in milliseconds Seerr waits for external API responses.",
				Optional:            true,
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"proxy": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"enabled":                schema.BoolAttribute{Optional: true, Computed: true},
					"hostname":               schema.StringAttribute{Optional: true, Computed: true},
					"port":                   schema.Int64Attribute{Optional: true, Computed: true},
					"use_ssl":                schema.BoolAttribute{Optional: true, Computed: true},
					"user":                   schema.StringAttribute{Optional: true, Computed: true},
					"password":               schema.StringAttribute{Optional: true, Computed: true, Sensitive: true},
					"bypass_filter":          schema.StringAttribute{Optional: true, Computed: true},
					"bypass_local_addresses": schema.BoolAttribute{Optional: true, Computed: true},
				},
			},
			"dns_cache": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"enabled":       schema.BoolAttribute{Optional: true, Computed: true},
					"force_min_ttl": schema.Int64Attribute{Optional: true, Computed: true},
					"force_max_ttl": schema.Int64Attribute{Optional: true, Computed: true},
				},
			},
		},
	}
}

func (r *NetworkSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = networkSettingsResourceSchema()
}

func (r *NetworkSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NetworkSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyNetworkSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Create Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkSettingsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.refreshNetworkSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkSettingsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyNetworkSettings(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Update Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkSettingsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Network settings are a singleton and remain managed remotely after Terraform state removal.
}

func (r *NetworkSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *NetworkSettingsResource) applyNetworkSettings(ctx context.Context, data *NetworkSettingsModel) error {
	payload, err := r.buildNetworkPayload(ctx, data)
	if err != nil {
		return err
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	res, err := r.client.Request(ctx, "POST", "/api/v1/settings/network", string(body), nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	return r.refreshNetworkSettings(ctx, data)
}

func (r *NetworkSettingsResource) refreshNetworkSettings(ctx context.Context, data *NetworkSettingsModel) error {
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/network", "", nil)
	if err != nil {
		return err
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var decoded map[string]any
	if err := json.Unmarshal(res.Body, &decoded); err != nil {
		return err
	}

	previous := *data
	r.applyNetworkSettingsMap(data, decoded)
	if previous.Proxy == nil {
		data.Proxy = nil
	}
	if data.Proxy != nil && previous.Proxy != nil && data.Proxy.Password.IsNull() && !previous.Proxy.Password.IsNull() && !previous.Proxy.Password.IsUnknown() {
		data.Proxy.Password = previous.Proxy.Password
	}
	if previous.DNSCache == nil {
		data.DNSCache = nil
	}
	data.ID = types.StringValue("network")
	return nil
}

func (r *NetworkSettingsResource) buildNetworkPayload(ctx context.Context, data *NetworkSettingsModel) (map[string]any, error) {
	res, err := r.client.Request(ctx, "GET", "/api/v1/settings/network", "", nil)
	if err != nil {
		return nil, err
	}
	if !StatusIsOK(res.StatusCode) {
		return nil, fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var payload map[string]any
	if err := json.Unmarshal(res.Body, &payload); err != nil {
		return nil, err
	}

	setOptionalBool(payload, "csrfProtection", data.CSRFProtection)
	setOptionalBool(payload, "forceIpv4First", data.ForceIPv4First)
	setOptionalBool(payload, "trustProxy", data.TrustProxy)
	setOptionalInt64(payload, "apiRequestTimeout", data.APIRequestTimeoutMS)

	if data.Proxy != nil {
		proxy := map[string]any{}
		if rawProxy, ok := payload["proxy"].(map[string]any); ok {
			proxy = copyMap(rawProxy)
		}
		setOptionalBool(proxy, "enabled", data.Proxy.Enabled)
		setOptionalString(proxy, "hostname", data.Proxy.Hostname)
		setOptionalInt64(proxy, "port", data.Proxy.Port)
		setOptionalBool(proxy, "useSsl", data.Proxy.UseSSL)
		setOptionalString(proxy, "user", data.Proxy.User)
		setOptionalString(proxy, "password", data.Proxy.Password)
		setOptionalString(proxy, "bypassFilter", data.Proxy.BypassFilter)
		setOptionalBool(proxy, "bypassLocalAddresses", data.Proxy.BypassLocalAddresses)
		payload["proxy"] = proxy
	}

	if data.DNSCache != nil {
		dnsCache := map[string]any{}
		if rawDNSCache, ok := payload["dnsCache"].(map[string]any); ok {
			dnsCache = copyMap(rawDNSCache)
		}
		setOptionalBool(dnsCache, "enabled", data.DNSCache.Enabled)
		setOptionalInt64(dnsCache, "forceMinTtl", data.DNSCache.ForceMinTTL)
		setOptionalInt64(dnsCache, "forceMaxTtl", data.DNSCache.ForceMaxTTL)
		payload["dnsCache"] = dnsCache
	}

	return payload, nil
}

func (r *NetworkSettingsResource) applyNetworkSettingsMap(data *NetworkSettingsModel, decoded map[string]any) {
	data.CSRFProtection = types.BoolNull()
	data.ForceIPv4First = types.BoolNull()
	data.TrustProxy = types.BoolNull()
	data.APIRequestTimeoutMS = types.Int64Null()
	data.Proxy = nil
	data.DNSCache = nil

	if v, ok := boolValueFromAny(decoded["csrfProtection"]); ok {
		data.CSRFProtection = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["forceIpv4First"]); ok {
		data.ForceIPv4First = types.BoolValue(v)
	}
	if v, ok := boolValueFromAny(decoded["trustProxy"]); ok {
		data.TrustProxy = types.BoolValue(v)
	}
	if v, ok := int64ValueFromAny(decoded["apiRequestTimeout"]); ok {
		data.APIRequestTimeoutMS = types.Int64Value(v)
	}

	if rawProxy, ok := decoded["proxy"].(map[string]any); ok {
		proxy := &NetworkProxyModel{
			Enabled:              types.BoolNull(),
			Hostname:             types.StringNull(),
			Port:                 types.Int64Null(),
			UseSSL:               types.BoolNull(),
			User:                 types.StringNull(),
			Password:             types.StringNull(),
			BypassFilter:         types.StringNull(),
			BypassLocalAddresses: types.BoolNull(),
		}
		if v, ok := boolValueFromAny(rawProxy["enabled"]); ok {
			proxy.Enabled = types.BoolValue(v)
		}
		if v, ok := stringValueFromAny(rawProxy["hostname"]); ok {
			proxy.Hostname = types.StringValue(v)
		}
		if v, ok := int64ValueFromAny(rawProxy["port"]); ok {
			proxy.Port = types.Int64Value(v)
		}
		if v, ok := boolValueFromAny(rawProxy["useSsl"]); ok {
			proxy.UseSSL = types.BoolValue(v)
		}
		if v, ok := stringValueFromAny(rawProxy["user"]); ok {
			proxy.User = types.StringValue(v)
		}
		if v, ok := stringValueFromAny(rawProxy["password"]); ok {
			proxy.Password = types.StringValue(v)
		}
		if v, ok := stringValueFromAny(rawProxy["bypassFilter"]); ok {
			proxy.BypassFilter = types.StringValue(v)
		}
		if v, ok := boolValueFromAny(rawProxy["bypassLocalAddresses"]); ok {
			proxy.BypassLocalAddresses = types.BoolValue(v)
		}
		data.Proxy = proxy
	}

	if rawDNSCache, ok := decoded["dnsCache"].(map[string]any); ok {
		dnsCache := &NetworkDNSCacheModel{
			Enabled:     types.BoolNull(),
			ForceMinTTL: types.Int64Null(),
			ForceMaxTTL: types.Int64Null(),
		}
		if v, ok := boolValueFromAny(rawDNSCache["enabled"]); ok {
			dnsCache.Enabled = types.BoolValue(v)
		}
		if v, ok := int64ValueFromAny(rawDNSCache["forceMinTtl"]); ok {
			dnsCache.ForceMinTTL = types.Int64Value(v)
		}
		if v, ok := int64ValueFromAny(rawDNSCache["forceMaxTtl"]); ok {
			dnsCache.ForceMaxTTL = types.Int64Value(v)
		}
		data.DNSCache = dnsCache
	}
}
