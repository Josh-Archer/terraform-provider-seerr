package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &PlexDevicesDataSource{}

type PlexDevicesDataSource struct {
	client *APIClient
}

type PlexConnectionModel struct {
	Protocol types.String `tfsdk:"protocol"`
	Address  types.String `tfsdk:"address"`
	Port     types.Int64  `tfsdk:"port"`
	URI      types.String `tfsdk:"uri"`
	Local    types.Bool   `tfsdk:"local"`
	Status   types.Int64  `tfsdk:"status"`
	Message  types.String `tfsdk:"message"`
}

type PlexDeviceModel struct {
	Name                   types.String          `tfsdk:"name"`
	Product                types.String          `tfsdk:"product"`
	ProductVersion         types.String          `tfsdk:"product_version"`
	Platform               types.String          `tfsdk:"platform"`
	PlatformVersion        types.String          `tfsdk:"platform_version"`
	Device                 types.String          `tfsdk:"device"`
	ClientIdentifier       types.String          `tfsdk:"client_identifier"`
	CreatedAt              types.String          `tfsdk:"created_at"`
	LastSeenAt             types.String          `tfsdk:"last_seen_at"`
	Provides               []types.String        `tfsdk:"provides"`
	Owned                  types.Bool            `tfsdk:"owned"`
	OwnerID                types.String          `tfsdk:"owner_id"`
	Home                   types.Bool            `tfsdk:"home"`
	SourceTitle            types.String          `tfsdk:"source_title"`
	AccessToken            types.String          `tfsdk:"access_token"`
	PublicAddress          types.String          `tfsdk:"public_address"`
	HTTPSRequired          types.Bool            `tfsdk:"https_required"`
	Synced                 types.Bool            `tfsdk:"synced"`
	Relay                  types.Bool            `tfsdk:"relay"`
	DNSRebindingProtection types.Bool            `tfsdk:"dns_rebinding_protection"`
	NATLoopbackSupported   types.Bool            `tfsdk:"nat_loopback_supported"`
	PublicAddressMatches   types.Bool            `tfsdk:"public_address_matches"`
	Presence               types.Bool            `tfsdk:"presence"`
	Connection             []PlexConnectionModel `tfsdk:"connection"`
}

type PlexDevicesDataSourceModel struct {
	ID      types.String      `tfsdk:"id"`
	Devices []PlexDeviceModel `tfsdk:"devices"`
}

func NewPlexDevicesDataSource() datasource.DataSource {
	return &PlexDevicesDataSource{}
}

func (d *PlexDevicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_plex_devices"
}

func (d *PlexDevicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about all available Plex server devices.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder ID for the data source.",
				Computed:            true,
			},
			"devices": schema.ListNestedAttribute{
				MarkdownDescription: "List of Plex devices.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the Plex device.",
							Computed:            true,
						},
						"product": schema.StringAttribute{
							MarkdownDescription: "Product name.",
							Computed:            true,
						},
						"product_version": schema.StringAttribute{
							MarkdownDescription: "Product version.",
							Computed:            true,
						},
						"platform": schema.StringAttribute{
							MarkdownDescription: "Platform of the device.",
							Computed:            true,
						},
						"platform_version": schema.StringAttribute{
							MarkdownDescription: "Platform version.",
							Computed:            true,
						},
						"device": schema.StringAttribute{
							MarkdownDescription: "Device type.",
							Computed:            true,
						},
						"client_identifier": schema.StringAttribute{
							MarkdownDescription: "Client identifier.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Creation time.",
							Computed:            true,
						},
						"last_seen_at": schema.StringAttribute{
							MarkdownDescription: "Last seen time.",
							Computed:            true,
						},
						"provides": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "Services provided by the device.",
							Computed:            true,
						},
						"owned": schema.BoolAttribute{
							MarkdownDescription: "Whether the device is owned.",
							Computed:            true,
						},
						"owner_id": schema.StringAttribute{
							MarkdownDescription: "Owner ID.",
							Computed:            true,
						},
						"home": schema.BoolAttribute{
							MarkdownDescription: "Whether the device is a home device.",
							Computed:            true,
						},
						"source_title": schema.StringAttribute{
							MarkdownDescription: "Source title.",
							Computed:            true,
						},
						"access_token": schema.StringAttribute{
							MarkdownDescription: "Access token for the device.",
							Computed:            true,
						},
						"public_address": schema.StringAttribute{
							MarkdownDescription: "Public address.",
							Computed:            true,
						},
						"https_required": schema.BoolAttribute{
							MarkdownDescription: "Whether HTTPS is required.",
							Computed:            true,
						},
						"synced": schema.BoolAttribute{
							MarkdownDescription: "Whether the device is synced.",
							Computed:            true,
						},
						"relay": schema.BoolAttribute{
							MarkdownDescription: "Whether the device uses a relay.",
							Computed:            true,
						},
						"dns_rebinding_protection": schema.BoolAttribute{
							MarkdownDescription: "Whether DNS rebinding protection is enabled.",
							Computed:            true,
						},
						"nat_loopback_supported": schema.BoolAttribute{
							MarkdownDescription: "Whether NAT loopback is supported.",
							Computed:            true,
						},
						"public_address_matches": schema.BoolAttribute{
							MarkdownDescription: "Whether the public address matches.",
							Computed:            true,
						},
						"presence": schema.BoolAttribute{
							MarkdownDescription: "Presence status.",
							Computed:            true,
						},
						"connection": schema.ListNestedAttribute{
							MarkdownDescription: "Connection details.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"protocol": schema.StringAttribute{
										MarkdownDescription: "Protocol used.",
										Computed:            true,
									},
									"address": schema.StringAttribute{
										MarkdownDescription: "Address of the connection.",
										Computed:            true,
									},
									"port": schema.Int64Attribute{
										MarkdownDescription: "Port used.",
										Computed:            true,
									},
									"uri": schema.StringAttribute{
										MarkdownDescription: "URI of the connection.",
										Computed:            true,
									},
									"local": schema.BoolAttribute{
										MarkdownDescription: "Whether the connection is local.",
										Computed:            true,
									},
									"status": schema.Int64Attribute{
										MarkdownDescription: "Status of the connection.",
										Computed:            true,
									},
									"message": schema.StringAttribute{
										MarkdownDescription: "Message from the connection.",
										Computed:            true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *PlexDevicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PlexDevicesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PlexDevicesDataSourceModel

	res, err := d.client.Request(ctx, "GET", "/api/v1/settings/plex/devices/servers", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !HandleAPIResponse(ctx, res, &resp.Diagnostics, "Read") {
		return
	}

	var devices []map[string]any
	if err := json.Unmarshal(res.Body, &devices); err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	data.Devices = make([]PlexDeviceModel, 0, len(devices))
	for _, dev := range devices {
		device := PlexDeviceModel{}

		if v, ok := dev["name"].(string); ok {
			device.Name = types.StringValue(v)
		}
		if v, ok := dev["product"].(string); ok {
			device.Product = types.StringValue(v)
		}
		if v, ok := dev["productVersion"].(string); ok {
			device.ProductVersion = types.StringValue(v)
		}
		if v, ok := dev["platform"].(string); ok {
			device.Platform = types.StringValue(v)
		}
		if v, ok := dev["platformVersion"].(string); ok {
			device.PlatformVersion = types.StringValue(v)
		}
		if v, ok := dev["device"].(string); ok {
			device.Device = types.StringValue(v)
		}
		if v, ok := dev["clientIdentifier"].(string); ok {
			device.ClientIdentifier = types.StringValue(v)
		}
		if v, ok := dev["createdAt"].(string); ok {
			device.CreatedAt = types.StringValue(v)
		}
		if v, ok := dev["lastSeenAt"].(string); ok {
			device.LastSeenAt = types.StringValue(v)
		}

		if provs, ok := dev["provides"].([]any); ok {
			device.Provides = make([]types.String, 0, len(provs))
			for _, p := range provs {
				if ps, ok := p.(string); ok {
					device.Provides = append(device.Provides, types.StringValue(ps))
				}
			}
		}

		if v, ok := dev["owned"].(bool); ok {
			device.Owned = types.BoolValue(v)
		}
		if v, ok := dev["ownerID"].(string); ok {
			device.OwnerID = types.StringValue(v)
		} else if v, ok := dev["ownerId"].(string); ok {
			device.OwnerID = types.StringValue(v)
		}
		if v, ok := dev["home"].(bool); ok {
			device.Home = types.BoolValue(v)
		}
		if v, ok := dev["sourceTitle"].(string); ok {
			device.SourceTitle = types.StringValue(v)
		}
		if v, ok := dev["accessToken"].(string); ok {
			device.AccessToken = types.StringValue(v)
		}
		if v, ok := dev["publicAddress"].(string); ok {
			device.PublicAddress = types.StringValue(v)
		}
		if v, ok := dev["httpsRequired"].(bool); ok {
			device.HTTPSRequired = types.BoolValue(v)
		}
		if v, ok := dev["synced"].(bool); ok {
			device.Synced = types.BoolValue(v)
		}
		if v, ok := dev["relay"].(bool); ok {
			device.Relay = types.BoolValue(v)
		}
		if v, ok := dev["dnsRebindingProtection"].(bool); ok {
			device.DNSRebindingProtection = types.BoolValue(v)
		}
		if v, ok := dev["natLoopbackSupported"].(bool); ok {
			device.NATLoopbackSupported = types.BoolValue(v)
		}
		if v, ok := dev["publicAddressMatches"].(bool); ok {
			device.PublicAddressMatches = types.BoolValue(v)
		}
		if v, ok := dev["presence"].(bool); ok {
			device.Presence = types.BoolValue(v)
		}

		if conns, ok := dev["connection"].([]any); ok {
			device.Connection = make([]PlexConnectionModel, 0, len(conns))
			for _, c := range conns {
				if conn, ok := c.(map[string]any); ok {
					cm := PlexConnectionModel{}
					if v, ok := conn["protocol"].(string); ok {
						cm.Protocol = types.StringValue(v)
					}
					if v, ok := conn["address"].(string); ok {
						cm.Address = types.StringValue(v)
					}
					if v, ok := conn["port"].(float64); ok {
						cm.Port = types.Int64Value(int64(v))
					}
					if v, ok := conn["uri"].(string); ok {
						cm.URI = types.StringValue(v)
					}
					if v, ok := conn["local"].(bool); ok {
						cm.Local = types.BoolValue(v)
					}
					if v, ok := conn["status"].(float64); ok {
						cm.Status = types.Int64Value(int64(v))
					}
					if v, ok := conn["message"].(string); ok {
						cm.Message = types.StringValue(v)
					}
					device.Connection = append(device.Connection, cm)
				}
			}
		}

		data.Devices = append(data.Devices, device)
	}

	data.ID = types.StringValue("plex_devices")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
