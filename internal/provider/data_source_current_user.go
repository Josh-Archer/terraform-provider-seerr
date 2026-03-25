package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CurrentUserDataSource{}

type CurrentUserDataSource struct {
	client *APIClient
}

type CurrentUserDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Email       types.String `tfsdk:"email"`
	Username    types.String `tfsdk:"username"`
	Permissions types.Int64  `tfsdk:"permissions"`
}

func NewCurrentUserDataSource() datasource.DataSource {
	return &CurrentUserDataSource{}
}

func (d *CurrentUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_current_user"
}

func (d *CurrentUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about the currently authenticated Seerr user via the `/api/v1/auth/me` endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the current user.",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address of the current user.",
				Computed:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username of the current user.",
				Computed:            true,
			},
			"permissions": schema.Int64Attribute{
				MarkdownDescription: "Permissions bitmask for the current user.",
				Computed:            true,
			},
		},
	}
}

func (d *CurrentUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CurrentUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CurrentUserDataSourceModel

	res, err := d.client.Request(ctx, "GET", "/api/v1/auth/me", "", nil)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}
	if !StatusIsOK(res.StatusCode) {
		resp.Diagnostics.AddError("Read Failed", fmt.Sprintf("status %d: %s", res.StatusCode, string(res.Body)))
		return
	}

	var user map[string]any
	if err := json.Unmarshal(res.Body, &user); err != nil {
		resp.Diagnostics.AddError("Read Failed", "Failed to parse API response: "+err.Error())
		return
	}

	// Populate Data Source
	idRaw := user["id"]
	switch v := idRaw.(type) {
	case float64:
		data.ID = types.StringValue(fmt.Sprintf("%.0f", v))
	case string:
		data.ID = types.StringValue(v)
	default:
		resp.Diagnostics.AddError("Parse Error", "Could not parse user ID.")
		return
	}

	if e, ok := user["email"].(string); ok {
		data.Email = types.StringValue(e)
	}
	if un, ok := user["username"].(string); ok {
		data.Username = types.StringValue(un)
	}
	if p, ok := user["permissions"].(float64); ok {
		data.Permissions = types.Int64Value(int64(p))
	} else if pStr, ok := user["permissions"].(string); ok {
		if pInt, err := strconv.ParseInt(pStr, 10, 64); err == nil {
			data.Permissions = types.Int64Value(pInt)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
