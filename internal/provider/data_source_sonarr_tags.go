package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SonarrTagsDataSource{}

type SonarrTagsDataSource struct{}

type SonarrTagModel struct {
	ID    types.Int64  `tfsdk:"id"`
	Label types.String `tfsdk:"label"`
}

type SonarrTagsDataSourceModel struct {
	URL      types.String     `tfsdk:"url"`
	Hostname types.String     `tfsdk:"hostname"`
	Port     types.Int64      `tfsdk:"port"`
	APIKey   types.String     `tfsdk:"api_key"`
	UseSSL   types.Bool       `tfsdk:"use_ssl"`
	BaseURL  types.String     `tfsdk:"base_url"`
	Tags     []SonarrTagModel `tfsdk:"tags"`
}

func NewSonarrTagsDataSource() datasource.DataSource {
	return &SonarrTagsDataSource{}
}

func (d *SonarrTagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sonarr_tags"
}

func (d *SonarrTagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch Sonarr tags via /api/v3/tag.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Full Sonarr URL. If set, it takes precedence over hostname, port, use_ssl, and base_url.",
				Optional:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Sonarr hostname when url is not provided.",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Sonarr port when url is not provided.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "Sonarr API key.",
				Required:            true,
				Sensitive:           true,
			},
			"use_ssl": schema.BoolAttribute{
				MarkdownDescription: "Use HTTPS when url is not provided.",
				Optional:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "Base path when Sonarr is served under a subpath.",
				Optional:            true,
			},
			"tags": schema.ListNestedAttribute{
				MarkdownDescription: "List of tags from Sonarr.",
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

func (d *SonarrTagsDataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
}

func (d *SonarrTagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SonarrTagsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostname := data.Hostname.ValueString()
	if data.Hostname.IsNull() || data.Hostname.IsUnknown() || strings.TrimSpace(hostname) == "" {
		hostname = "sonarr-service"
	}
	port := data.Port.ValueInt64()
	if data.Port.IsNull() || data.Port.IsUnknown() || port == 0 {
		port = 8989
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

	var tags []SonarrTagModel
	for _, item := range results {
		tag := SonarrTagModel{}
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
