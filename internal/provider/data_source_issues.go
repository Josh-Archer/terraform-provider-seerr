package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IssuesDataSource{}

type IssuesDataSource struct {
	client *APIClient
}

type IssueSummaryModel struct {
	ID          types.String `tfsdk:"id"`
	IssueType   types.Int64  `tfsdk:"issue_type"`
	Status      types.Int64  `tfsdk:"status"`
	MediaID     types.Int64  `tfsdk:"media_id"`
	CreatedByID types.Int64  `tfsdk:"created_by_id"`
}

type IssuesDataSourceModel struct {
	ID     types.String        `tfsdk:"id"`
	Issues []IssueSummaryModel `tfsdk:"issues"`
}

func NewIssuesDataSource() datasource.DataSource {
	return &IssuesDataSource{}
}

func (d *IssuesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_issues"
}

func (d *IssuesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about all existing Seerr issues.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"issues": schema.ListNestedAttribute{
				MarkdownDescription: "List of issues.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The ID of the issue.",
							Computed:            true,
						},
						"issue_type": schema.Int64Attribute{
							MarkdownDescription: "The type of the issue (1: Video, 2: Audio, 3: Subtitle, 4: Other, 5: Unknown).",
							Computed:            true,
						},
						"status": schema.Int64Attribute{
							MarkdownDescription: "The status of the issue (1: Open, 2: Resolved).",
							Computed:            true,
						},
						"media_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the media associated with the issue.",
							Computed:            true,
						},
						"created_by_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the user who created the issue.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *IssuesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IssuesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IssuesDataSourceModel

	// Fetch issues
	res, err := d.client.Request(ctx, "GET", "/api/v1/issue?take=1000", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var parsedResponse struct {
		Results []map[string]any `json:"results"`
	}
	if err := json.Unmarshal(res.Body, &parsedResponse); err != nil {
		resp.Diagnostics.AddError("Read Failed", "Failed to parse API response: "+err.Error())
		return
	}

	for _, u := range parsedResponse.Results {
		issue := IssueSummaryModel{}

		idRaw := u["id"]
		switch v := idRaw.(type) {
		case float64:
			issue.ID = types.StringValue(fmt.Sprintf("%.0f", v))
		case string:
			issue.ID = types.StringValue(v)
		}

		if it, ok := u["issueType"].(float64); ok {
			issue.IssueType = types.Int64Value(int64(it))
		}
		if s, ok := u["status"].(float64); ok {
			issue.Status = types.Int64Value(int64(s))
		}

		if mediaRaw, ok := u["media"].(map[string]any); ok {
			if mediaId, ok := mediaRaw["id"].(float64); ok {
				issue.MediaID = types.Int64Value(int64(mediaId))
			}
		}

		if createdByRaw, ok := u["createdBy"].(map[string]any); ok {
			if createdById, ok := createdByRaw["id"].(float64); ok {
				issue.CreatedByID = types.Int64Value(int64(createdById))
			}
		}

		data.Issues = append(data.Issues, issue)
	}

	data.ID = types.StringValue("all_issues")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
