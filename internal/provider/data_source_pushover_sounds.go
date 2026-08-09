package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PushoverSoundsDataSource{}

type PushoverSoundsDataSource struct {
	client *APIClient
}

type PushoverSoundModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

type PushoverSoundsDataSourceModel struct {
	ID     types.String         `tfsdk:"id"`
	Token  types.String         `tfsdk:"token"`
	Sounds []PushoverSoundModel `tfsdk:"sounds"`
}

func NewPushoverSoundsDataSource() datasource.DataSource {
	return &PushoverSoundsDataSource{}
}

func (d *PushoverSoundsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pushover_sounds"
}

func (d *PushoverSoundsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get all available Pushover notification sounds from Seerr.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "The Pushover application API token. Required to query available sounds.",
				Required:            true,
				Sensitive:           true,
			},
			"sounds": schema.ListNestedAttribute{
				MarkdownDescription: "List of available Pushover sounds.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The identifier name of the sound.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "The human-readable description of the sound.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *PushoverSoundsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PushoverSoundsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PushoverSoundsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Make the API request with the required token query parameter
	urlPath := fmt.Sprintf("/api/v1/settings/notifications/pushover/sounds?token=%s", url.QueryEscape(data.Token.ValueString()))
	res, err := d.client.Request(ctx, "GET", urlPath, "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var parsedSounds []map[string]any
	if err := json.Unmarshal(res.Body, &parsedSounds); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	var sounds []PushoverSoundModel
	for _, s := range parsedSounds {
		sound := PushoverSoundModel{
			Name:        types.StringNull(),
			Description: types.StringNull(),
		}
		if name, ok := s["name"].(string); ok {
			sound.Name = types.StringValue(name)
		}
		if desc, ok := s["description"].(string); ok {
			sound.Description = types.StringValue(desc)
		}
		sounds = append(sounds, sound)
	}

	data.Sounds = sounds
	data.ID = types.StringValue("pushover_sounds")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
