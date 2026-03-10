package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NotificationAgentsDataSource{}

type NotificationAgentsDataSource struct {
	client *APIClient
}

type NotificationAgentSummaryModel struct {
	Agent       types.String `tfsdk:"agent"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	EmbedPoster types.Bool   `tfsdk:"embed_poster"`
}

type NotificationAgentsDataSourceModel struct {
	ID     types.String                    `tfsdk:"id"`
	Agents []NotificationAgentSummaryModel `tfsdk:"agents"`
}

func NewNotificationAgentsDataSource() datasource.DataSource {
	return &NotificationAgentsDataSource{}
}

func (d *NotificationAgentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_agents"
}

func (d *NotificationAgentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about all configured notification agents in Seerr.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"agents": schema.ListNestedAttribute{
				MarkdownDescription: "List of notification agents.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"agent": schema.StringAttribute{
							MarkdownDescription: "The name of the agent (e.g., discord, slack, email).",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the agent is enabled.",
							Computed:            true,
						},
						"embed_poster": schema.BoolAttribute{
							MarkdownDescription: "Whether to embed poster images in notifications.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *NotificationAgentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NotificationAgentsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NotificationAgentsDataSourceModel

	// Fetch notification settings
	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/notifications", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var settings map[string]any
	if err := json.Unmarshal(res.Body, &settings); err != nil {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("failed to decode response: %s", err))
		return
	}

	// The API returns a map where keys are agent names
	for agentName, agentData := range settings {
		if agentMap, ok := agentData.(map[string]any); ok {
			agent := NotificationAgentSummaryModel{
				Agent: types.StringValue(agentName),
			}

			if enabled, ok := agentMap["enabled"].(bool); ok {
				agent.Enabled = types.BoolValue(enabled)
			}
			if embed, ok := agentMap["embedPoster"].(bool); ok {
				agent.EmbedPoster = types.BoolValue(embed)
			}

			data.Agents = append(data.Agents, agent)
		}
	}

	data.ID = types.StringValue("all_notification_agents")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
