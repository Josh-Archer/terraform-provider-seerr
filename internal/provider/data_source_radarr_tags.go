package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RadarrTagsDataSource{}

type RadarrTagsDataSource struct{}

type RadarrTagModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Label types.String `tfsdk:"label"`
}

type RadarrTagsDataSourceModel struct {
	URL      types.String     `tfsdk:"url"`
	Hostname types.String     `tfsdk:"hostname"`
	Port     types.Int64      `tfsdk:"port"`
	APIKey   types.String     `tfsdk:"api_key"`
	UseSSL   types.Bool       `tfsdk:"use_ssl"`
	BaseURL  types.String     `tfsdk:"base_url"`
	Tags     []RadarrTagModel `tfsdk:"tags"`
}

func NewRadarrTagsDataSource() datasource.DataSource {
	return &RadarrTagsDataSource{}
}

func (d *RadarrTagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radarr_tags"
}

func (d *RadarrTagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch Radarr tags via /api/v3/tag.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Full Radarr URL. If set, it takes precedence over hostname, port, use_ssl, and base_url.",
				Optional:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Radarr hostname when url is not provided.",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Radarr port when url is not provided.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Radarr API key.",
				Required:            true,
				Sensitive:           true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Use HTTPS when url is not provided.",
				Optional:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base path when Radarr is served under a subpath.",
				Optional:            true,
			},
			"tags": schema.ListNestedAttribute{
				MarkdownDescription: "List of tags from Radarr.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Tag ID.",
							Computed:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "Tag label.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *RadarrTagsDataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
}

func (d *RadarrTagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RadarrTagsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostname := data.Hostname.ValueString()
	if data.Hostname.IsNull() || data.Hostname.IsUnknown() || strings.TrimSpace(hostname) == "" {
		hostname = "radarr-service"
	}
	port := data.Port.ValueInt64()
	if data.Port.IsNull() || data.Port.IsUnknown() || port == 0 {
		port = 7878
	}
	useSSL := false
	if !data.UseSSL.IsNull() && !data.UseSSL.IsUnknown() {
		useSSL = data.UseSSL.ValueBool()
	}
	baseURL := ""
	if !data.BaseURL.IsNull() && !data.BaseURL.IsUnknown() {
		baseURL = data.BaseURL.ValueString()
	}

	results, err := fetchArrEndpoint(
		ctx,
		data.URL.ValueString(),
		hostname,
		port,
		useSSL,
		baseURL,
		data.APIKey.ValueString(),
		"/api/v3/tag",
	)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	tags := make([]RadarrTagModel, 0, len(results))
	for _, item := range results {
		tag := RadarrTagModel{}
		if id, ok := int64ValueFromAny(item["id"]); ok {
			tag.ID = types.Int64Value(id)
		}
		if label, ok := stringValueFromAny(item["label"]); ok {
			tag.Label = types.StringValue(label)
		}
		tags = append(tags, tag)
	}

	data.Tags = tags
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
