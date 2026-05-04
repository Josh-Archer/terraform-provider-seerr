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

var _ datasource.DataSource = &MediaDataSource{}

type MediaDataSource struct {
	client *APIClient
}

type MediaDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	Filter       types.String `tfsdk:"filter"`
	Sort         types.String `tfsdk:"sort"`
	Take         types.Int64  `tfsdk:"take"`
	Skip         types.Int64  `tfsdk:"skip"`
	ResultsJSON  types.String `tfsdk:"results_json"`
	PageInfoJSON types.String `tfsdk:"page_info_json"`
	ResponseJSON types.String `tfsdk:"response_json"`
}

func NewMediaDataSource() datasource.DataSource { return &MediaDataSource{} }

func (d *MediaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_media"
}

func (d *MediaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read Seerr media records via `/api/v1/media`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"filter": schema.StringAttribute{
				MarkdownDescription: "Optional media status filter: `available`, `partial`, `allavailable`, `processing`, or `pending`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("available", "partial", "allavailable", "processing", "pending"),
				},
			},
			"sort": schema.StringAttribute{
				MarkdownDescription: "Optional media sort: `modified` or `mediaAdded`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("modified", "mediaAdded"),
				},
			},
			"take":           schema.Int64Attribute{Optional: true},
			"skip":           schema.Int64Attribute{Optional: true},
			"results_json":   schema.StringAttribute{Computed: true},
			"page_info_json": schema.StringAttribute{Computed: true},
			"response_json":  schema.StringAttribute{Computed: true},
		},
	}
}

func (d *MediaDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *MediaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MediaDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	endpoint := "/api/v1/media"
	values := url.Values{}
	if !data.Filter.IsNull() && !data.Filter.IsUnknown() && data.Filter.ValueString() != "" {
		values.Set("filter", data.Filter.ValueString())
	}
	if !data.Sort.IsNull() && !data.Sort.IsUnknown() && data.Sort.ValueString() != "" {
		values.Set("sort", data.Sort.ValueString())
	}
	if !data.Take.IsNull() && !data.Take.IsUnknown() {
		values.Set("take", strconv.FormatInt(data.Take.ValueInt64(), 10))
	}
	if !data.Skip.IsNull() && !data.Skip.IsUnknown() {
		values.Set("skip", strconv.FormatInt(data.Skip.ValueInt64(), 10))
	}
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
	data.ID = types.StringValue("media")
	data.ResponseJSON = types.StringValue(string(res.Body))
	if raw, ok := jsonFieldString(res.Body, "results"); ok {
		data.ResultsJSON = types.StringValue(raw)
	}
	if raw, ok := jsonFieldString(res.Body, "pageInfo"); ok {
		data.PageInfoJSON = types.StringValue(raw)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
