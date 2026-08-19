package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &RadarrRootFoldersDataSource{}

type RadarrRootFoldersDataSource struct{}

type RadarrRootFolderModel struct {
	ID         types.Int64  `tfsdk:"id"`
	Path       types.String `tfsdk:"path"`
	Accessible types.Bool   `tfsdk:"accessible"`
	FreeSpace  types.Int64  `tfsdk:"free_space"`
}

type RadarrRootFoldersDataSourceModel struct {
	URL         types.String            `tfsdk:"url"`
	Hostname    types.String            `tfsdk:"hostname"`
	Port        types.Int64             `tfsdk:"port"`
	APIKey      types.String            `tfsdk:"api_key"`
	UseSSL      types.Bool              `tfsdk:"use_ssl"`
	BaseURL     types.String            `tfsdk:"base_url"`
	RootFolders []RadarrRootFolderModel `tfsdk:"root_folders"`
}

func NewRadarrRootFoldersDataSource() datasource.DataSource {
	return &RadarrRootFoldersDataSource{}
}

func (d *RadarrRootFoldersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_radarr_root_folders"
}

func (d *RadarrRootFoldersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch Radarr root folders via /api/v3/rootfolder.",
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
			"root_folders": schema.ListNestedAttribute{
				MarkdownDescription: "List of root folders from Radarr.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Root folder ID.",
							Computed:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "Root folder path.",
							Computed:            true,
						},
						"accessible": schema.BoolAttribute{
							MarkdownDescription: "Whether the root folder is accessible.",
							Computed:            true,
						},
						"free_space": schema.Int64Attribute{
							MarkdownDescription: "Free space in bytes.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *RadarrRootFoldersDataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
}

func (d *RadarrRootFoldersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RadarrRootFoldersDataSourceModel
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
		defaultRequestTimeout,
		"/api/v3/rootfolder",
	)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	var rootFolders []RadarrRootFolderModel
	for _, item := range results {
		folder := RadarrRootFolderModel{}
		if id, ok := int64ValueFromAny(item["id"]); ok {
			folder.ID = types.Int64Value(id)
		}
		if path, ok := stringValueFromAny(item["path"]); ok {
			folder.Path = types.StringValue(path)
		}
		if accessible, ok := boolValueFromAny(item["accessible"]); ok {
			folder.Accessible = types.BoolValue(accessible)
		}
		if freeSpace, ok := int64ValueFromAny(item["freeSpace"]); ok {
			folder.FreeSpace = types.Int64Value(freeSpace)
		}
		rootFolders = append(rootFolders, folder)
	}

	data.RootFolders = rootFolders
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
