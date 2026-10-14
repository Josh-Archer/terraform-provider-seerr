package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ServiceServerSummaryModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	IsDefault       types.Bool   `tfsdk:"is_default"`
	ActiveDirectory types.String `tfsdk:"active_directory"`
	Is4K            types.Bool   `tfsdk:"is_4k"`
}

type ServiceProfileModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func serviceDataSourceSchema(serviceName string) schema.Schema {
	title := serviceName
	return schema.Schema{
		MarkdownDescription: fmt.Sprintf("Read %s service reachability, server summary information, and quality profiles from Seerr.", title),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: fmt.Sprintf("The identifier for this %s service query.", title),
				Computed:            true,
			},
			"server_id": schema.Int64Attribute{
				MarkdownDescription: fmt.Sprintf("Optional %s server ID to query. If omitted, all %s servers are queried.", title, title),
				Optional:            true,
			},
			"servers": schema.ListNestedAttribute{
				MarkdownDescription: fmt.Sprintf("List of reachable %s servers.", title),
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: fmt.Sprintf("%s server ID.", title),
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: fmt.Sprintf("%s server name.", title),
							Computed:            true,
						},
						"is_default": schema.BoolAttribute{
							MarkdownDescription: "Whether this is the default server.",
							Computed:            true,
						},
						"active_directory": schema.StringAttribute{
							MarkdownDescription: "Active directory for the server.",
							Computed:            true,
						},
						"is_4k": schema.BoolAttribute{
							MarkdownDescription: "Whether the server is configured for 4K.",
							Computed:            true,
						},
					},
				},
			},
			"profiles": schema.ListNestedAttribute{
				MarkdownDescription: fmt.Sprintf("List of quality profiles when querying a specific %s server_id.", title),
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Quality profile ID.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Quality profile name.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func parseServiceServerSummary(m map[string]any) ServiceServerSummaryModel {
	var s ServiceServerSummaryModel

	if v, ok := int64ValueFromAny(m["id"]); ok {
		s.ID = types.Int64Value(v)
	}
	if v, ok := stringValueFromAny(m["name"]); ok {
		s.Name = types.StringValue(v)
	}

	for _, k := range []string{"isDefault", "is_default"} {
		if val, ok := m[k]; ok {
			if b, ok := boolValueFromAny(val); ok {
				s.IsDefault = types.BoolValue(b)
				break
			} else if val == nil {
				s.IsDefault = types.BoolNull()
				break
			}
		}
	}

	for _, k := range []string{"activeDirectory", "active_directory"} {
		if val, ok := m[k]; ok {
			if str, ok := stringValueFromAny(val); ok {
				s.ActiveDirectory = types.StringValue(str)
				break
			} else if val == nil {
				s.ActiveDirectory = types.StringNull()
				break
			}
		}
	}

	for _, k := range []string{"is4k", "is_4k"} {
		if val, ok := m[k]; ok {
			if b, ok := boolValueFromAny(val); ok {
				s.Is4K = types.BoolValue(b)
				break
			} else if val == nil {
				s.Is4K = types.BoolNull()
				break
			}
		}
	}

	return s
}

func parseServiceProfiles(raw any) []ServiceProfileModel {
	if raw == nil {
		return []ServiceProfileModel{}
	}
	switch v := raw.(type) {
	case []any:
		profiles := make([]ServiceProfileModel, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				profiles = append(profiles, parseSingleProfile(m))
			}
		}
		return profiles
	case map[string]any:
		return []ServiceProfileModel{parseSingleProfile(v)}
	default:
		return []ServiceProfileModel{}
	}
}

func parseSingleProfile(m map[string]any) ServiceProfileModel {
	var p ServiceProfileModel
	if v, ok := int64ValueFromAny(m["id"]); ok {
		p.ID = types.Int64Value(v)
	}
	if v, ok := stringValueFromAny(m["name"]); ok {
		p.Name = types.StringValue(v)
	}
	return p
}

func readServiceReachability(ctx context.Context, client *APIClient, serviceType string, serverID types.Int64, diags *diag.Diagnostics) (string, []ServiceServerSummaryModel, []ServiceProfileModel) {
	if !serverID.IsNull() && !serverID.IsUnknown() {
		endpoint := fmt.Sprintf("/api/v1/service/%s/%d", serviceType, serverID.ValueInt64())
		res, err := client.Request(ctx, "GET", endpoint, "", nil)
		if err != nil {
			diags.AddError("Read Failed", err.Error())
			return "", nil, nil
		}
		if !HandleAPIResponse(ctx, res, diags, "Read") {
			return "", nil, nil
		}

		var payload map[string]any
		if err := json.Unmarshal(res.Body, &payload); err != nil {
			diags.AddError("Read Failed", fmt.Sprintf("Failed to parse response: %s", err))
			return "", nil, nil
		}

		var servers []ServiceServerSummaryModel
		if srvRaw, ok := payload["server"].(map[string]any); ok {
			servers = []ServiceServerSummaryModel{parseServiceServerSummary(srvRaw)}
		} else {
			servers = []ServiceServerSummaryModel{parseServiceServerSummary(payload)}
		}

		profiles := parseServiceProfiles(payload["profiles"])
		id := fmt.Sprintf("%d", serverID.ValueInt64())
		return id, servers, profiles
	}

	endpoint := fmt.Sprintf("/api/v1/service/%s", serviceType)
	res, err := client.Request(ctx, "GET", endpoint, "", nil)
	if err != nil {
		diags.AddError("Read Failed", err.Error())
		return "", nil, nil
	}
	if !HandleAPIResponse(ctx, res, diags, "Read") {
		return "", nil, nil
	}

	var list []map[string]any
	if err := json.Unmarshal(res.Body, &list); err != nil {
		var obj map[string]any
		if err2 := json.Unmarshal(res.Body, &obj); err2 == nil {
			if rawList, ok := obj["servers"].([]any); ok {
				for _, item := range rawList {
					if m, ok := item.(map[string]any); ok {
						list = append(list, m)
					}
				}
			}
		}
		if list == nil {
			diags.AddError("Read Failed", fmt.Sprintf("Failed to parse response: %s", err))
			return "", nil, nil
		}
	}

	servers := make([]ServiceServerSummaryModel, 0, len(list))
	for _, item := range list {
		servers = append(servers, parseServiceServerSummary(item))
	}

	id := fmt.Sprintf("%s_service", serviceType)
	return id, servers, nil
}
