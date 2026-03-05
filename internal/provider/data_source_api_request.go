package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &APIRequestDataSource{}

type APIRequestDataSource struct {
	client *APIClient
}

type APIRequestDataSourceModel struct {
	Path             types.String `tfsdk:"path"`
	Method           types.String `tfsdk:"method"`
	Headers          types.Map    `tfsdk:"headers"`
	RequestBodyJSON  types.String `tfsdk:"request_body_json"`
	ResponseBodyJSON types.String `tfsdk:"response_body_json"`
	StatusCode       types.Int64  `tfsdk:"status_code"`
}

func NewAPIRequestDataSource() datasource.DataSource {
	return &APIRequestDataSource{}
}

func (d *APIRequestDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_request"
}

func (d *APIRequestDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Execute arbitrary Seerr API requests and return the response.",
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				MarkdownDescription: "Endpoint path or absolute URL.",
				Required:            true,
			},
			"method": schema.StringAttribute{
				MarkdownDescription: "HTTP method used for the request.",
				Optional:            true,
			},
			"headers": schema.MapAttribute{
				MarkdownDescription: "Extra request headers.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"request_body_json": schema.StringAttribute{
				MarkdownDescription: "Optional JSON request body.",
				Optional:            true,
			},
			"response_body_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body.",
				Computed:            true,
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "HTTP status code.",
				Computed:            true,
			},
		},
	}
}

func (d *APIRequestDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*APIClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Configure Type", fmt.Sprintf("Expected *APIClient, got %T", req.ProviderData))
		return
	}

	d.client = client
}

func (d *APIRequestDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APIRequestDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	headers, diags := mapFromTypesMap(ctx, data.Headers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	method := "GET"
	if !data.Method.IsNull() && !data.Method.IsUnknown() && strings.TrimSpace(data.Method.ValueString()) != "" {
		method = data.Method.ValueString()
	}

	httpResp, err := d.client.Request(ctx, method, data.Path.ValueString(), data.RequestBodyJSON.ValueString(), headers)
	if err != nil {
		resp.Diagnostics.AddError("Request Failed", err.Error())
		return
	}

	data.StatusCode = types.Int64Value(int64(httpResp.StatusCode))
	data.ResponseBodyJSON = types.StringValue(string(httpResp.Body))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
