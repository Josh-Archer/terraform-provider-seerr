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

	agents, err := d.readNotificationAgentSummaries(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.Agents = agents
	data.ID = types.StringValue("all_notification_agents")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func notificationAggregateAgents() []string {
	return []string{
		"discord",
		"slack",
		"email",
		"lunasea",
		"telegram",
		"pushbullet",
		"pushover",
		"ntfy",
		"webhook",
		"gotify",
		"webpush",
	}
}

func (d *NotificationAgentsDataSource) readNotificationAgentSummaries(ctx context.Context) ([]NotificationAgentSummaryModel, error) {
	agents := []NotificationAgentSummaryModel{}
	for _, agentName := range notificationAggregateAgents() {
		res, err := d.client.Request(ctx, "GET", notificationPath(agentName), "", nil)
		if err != nil {
			return nil, err
		}
		if res.StatusCode == 404 {
			continue
		}
		if !StatusIsOK(res.StatusCode) {
			return nil, fmt.Errorf("%s: status %d: %s", agentName, res.StatusCode, string(res.Body))
		}

		var agentMap map[string]any
		if err := json.Unmarshal(res.Body, &agentMap); err != nil {
			return nil, fmt.Errorf("%s: failed to decode response: %s", agentName, err)
		}

		agent := NotificationAgentSummaryModel{
			Agent:       types.StringValue(agentName),
			Enabled:     types.BoolNull(),
			EmbedPoster: types.BoolNull(),
		}
		if enabled, ok := boolValueFromAny(agentMap["enabled"]); ok {
			agent.Enabled = types.BoolValue(enabled)
		}
		if embed, ok := boolValueFromAny(agentMap["embedPoster"]); ok {
			agent.EmbedPoster = types.BoolValue(embed)
		}

		agents = append(agents, agent)
	}

	return agents, nil
}
