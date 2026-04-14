package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NetworkSettingsDataSource{}

type NetworkSettingsDataSource struct {
	client *APIClient
}

func NewNetworkSettingsDataSource() datasource.DataSource { return &NetworkSettingsDataSource{} }

func (d *NetworkSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_settings"
}

func (d *NetworkSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Seerr network settings via `/api/v1/settings/network`.",
		Attributes: map[string]schema.Attribute{
			"id":               schema.StringAttribute{Computed: true},
			"csrf_protection":  schema.BoolAttribute{Computed: true},
			"force_ipv4_first": schema.BoolAttribute{Computed: true},
			"trust_proxy":      schema.BoolAttribute{Computed: true},
			"api_request_timeout_ms": schema.Int64Attribute{
				MarkdownDescription: "Maximum time in milliseconds Seerr waits for external API responses.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"proxy": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"enabled":                schema.BoolAttribute{Computed: true},
					"hostname":               schema.StringAttribute{Computed: true},
					"port":                   schema.Int64Attribute{Computed: true},
					"use_ssl":                schema.BoolAttribute{Computed: true},
					"user":                   schema.StringAttribute{Computed: true},
					"password":               schema.StringAttribute{Computed: true, Sensitive: true},
					"bypass_filter":          schema.StringAttribute{Computed: true},
					"bypass_local_addresses": schema.BoolAttribute{Computed: true},
				},
			},
			"dns_cache": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"enabled":       schema.BoolAttribute{Computed: true},
					"force_min_ttl": schema.Int64Attribute{Computed: true},
					"force_max_ttl": schema.Int64Attribute{Computed: true},
				},
			},
		},
	}
}

func (d *NetworkSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}
	d.client = c
}

func (d *NetworkSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NetworkSettingsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/network", "", nil)
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

	resource := &NetworkSettingsResource{}
	resource.applyNetworkSettingsMap(&data, decoded)
	data.ID = types.StringValue("network")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
