package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IssueDataSource{}

type IssueDataSource struct {
	client *APIClient
}

type IssueDataSourceModel struct {
	ID           types.String `tfsdk:"id"`
	IssueType    types.Int64  `tfsdk:"issue_type"`
	Status       types.Int64  `tfsdk:"status"`
	MediaID      types.Int64  `tfsdk:"media_id"`
	CreatedByID  types.Int64  `tfsdk:"created_by_id"`
	ResponseJSON types.String `tfsdk:"response_json"`
}

func NewIssueDataSource() datasource.DataSource {
	return &IssueDataSource{}
}

func (d *IssueDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issue"
}

func (d *IssueDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up a single Seerr issue by ID via `GET /api/v1/issue/{id}`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The issue ID.",
				Required:            true,
			},
			"issue_type": schema.Int64Attribute{
				MarkdownDescription: "The type of the issue (1: Video, 2: Audio, 3: Subtitle, 4: Other).",
				Computed:            true,
			},
			"status": schema.Int64Attribute{
				MarkdownDescription: "The status of the issue (1: Open, 2: Resolved).",
				Computed:            true,
			},
			"media_id": schema.Int64Attribute{
				MarkdownDescription: "The Seerr-internal media ID associated with the issue.",
				Computed:            true,
			},
			"created_by_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the user who created the issue.",
				Computed:            true,
			},
			"response_json": schema.StringAttribute{
				MarkdownDescription: "Raw JSON response body from the API.",
				Computed:            true,
			},
		},
	}
}

func (d *IssueDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IssueDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IssueDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := d.refreshIssue(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *IssueDataSource) refreshIssue(ctx context.Context, data *IssueDataSourceModel) error {
	id := data.ID.ValueString()
	res, err := d.client.Request(ctx, "GET", "/api/v1/issue/"+id, "", nil)
	if err != nil {
		return err
	}
	if res.StatusCode == 404 {
		return fmt.Errorf("issue %q not found", id)
	}
	if !StatusIsOK(res.StatusCode) {
		return fmt.Errorf("status %d: %s", res.StatusCode, string(res.Body))
	}

	var m map[string]any
	if err := json.Unmarshal(res.Body, &m); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	data.ResponseJSON = types.StringValue(string(res.Body))

	if issueType, ok := int64ValueFromAny(m["issueType"]); ok {
		data.IssueType = types.Int64Value(issueType)
	}
	if status, ok := int64ValueFromAny(m["status"]); ok {
		data.Status = types.Int64Value(status)
	}
	if media, ok := m["media"].(map[string]any); ok {
		if mediaID, ok := int64ValueFromAny(media["id"]); ok {
			data.MediaID = types.Int64Value(mediaID)
		}
	}
	if createdBy, ok := m["createdBy"].(map[string]any); ok {
		if createdByID, ok := int64ValueFromAny(createdBy["id"]); ok {
			data.CreatedByID = types.Int64Value(createdByID)
		}
	}

	return nil
}
