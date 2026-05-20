package provider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DiscoverDataSource{}

type DiscoverDataSource struct {
	client *APIClient
}

type DiscoverDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	MediaType    types.String `tfsdk:"media_type"`
	Page         types.Int64  `tfsdk:"page"`
	SortBy       types.String `tfsdk:"sort_by"`
	Genre        types.String `tfsdk:"genre"`
	Keywords     types.String `tfsdk:"keywords"`
	Language     types.String `tfsdk:"language"`
	WatchRegion  types.String `tfsdk:"watch_region"`
	ResultsJSON  types.String `tfsdk:"results_json"`
	ResponseJSON types.String `tfsdk:"response_json"`
}

func NewDiscoverDataSource() datasource.DataSource { return &DiscoverDataSource{} }

func (d *DiscoverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_discover"
}

func (d *DiscoverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Seerr discovery results via `/api/v1/discover/movies` or `/api/v1/discover/tv`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"media_type": schema.StringAttribute{
				MarkdownDescription: "Discovery media type: `movie` or `tv`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("movie", "tv"),
				},
			},
			"page":          schema.Int64Attribute{Optional: true},
			"sort_by":       schema.StringAttribute{Optional: true},
			"genre":         schema.StringAttribute{Optional: true},
			"keywords":      schema.StringAttribute{Optional: true},
			"language":      schema.StringAttribute{Optional: true},
			"watch_region":  schema.StringAttribute{Optional: true},
			"results_json":  schema.StringAttribute{Computed: true},
			"response_json": schema.StringAttribute{Computed: true},
		},
	}
}

func (d *DiscoverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DiscoverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DiscoverDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	kind := data.MediaType.ValueString()
	endpoint := "/api/v1/discover/movies"
	if kind == "tv" {
		endpoint = "/api/v1/discover/tv"
	}
	values := url.Values{}
	if !data.Page.IsNull() && !data.Page.IsUnknown() {
		values.Set("page", strconv.FormatInt(data.Page.ValueInt64(), 10))
	}
	setQueryString(values, "sortBy", data.SortBy)
	setQueryString(values, "genre", data.Genre)
	setQueryString(values, "keywords", data.Keywords)
	setQueryString(values, "language", data.Language)
	setQueryString(values, "watchRegion", data.WatchRegion)
	if encoded := values.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	res, err := d.client.Request(ctx, "GET", endpoint, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}
	data.ID = types.StringValue("discover_" + kind)
	data.ResponseJSON = types.StringValue(string(res.Body))
	if raw, ok := jsonFieldString(res.Body, "results"); ok {
		data.ResultsJSON = types.StringValue(raw)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func setQueryString(values url.Values, name string, value types.String) {
	if !value.IsNull() && !value.IsUnknown() && value.ValueString() != "" {
		values.Set(name, value.ValueString())
	}
}
