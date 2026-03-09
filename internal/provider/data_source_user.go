package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserDataSource{}

type UserDataSource struct {
	client *APIClient
}

type UserDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Email       types.String `tfsdk:"email"`
	Username    types.String `tfsdk:"username"`
	Permissions types.Int64  `tfsdk:"permissions"`
}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

func (d *UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get information about an existing Seerr user by email or username.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the user.",
				Computed:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "The email address of the user to look up. Exactly one of `email` or `username` must be provided.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username of the user to look up. Exactly one of `email` or `username` must be provided.",
				Optional:            true,
			},
			"permissions": schema.Int64Attribute{
				MarkdownDescription: "Permissions bitmask for the user.",
				Computed:            true,
			},
		},
	}
}

func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasEmail := !data.Email.IsNull() && !data.Email.IsUnknown() && data.Email.ValueString() != ""
	hasUsername := !data.Username.IsNull() && !data.Username.IsUnknown() && data.Username.ValueString() != ""

	if hasEmail && hasUsername {
		resp.Diagnostics.AddError("Invalid Configuration", "Exactly one of 'email' or 'username' must be provided, got both.")
		return
	}
	if !hasEmail && !hasUsername {
		resp.Diagnostics.AddError("Invalid Configuration", "Exactly one of 'email' or 'username' must be provided, got neither.")
		return
	}

	searchEmail := ""
	if hasEmail {
		searchEmail = strings.ToLower(data.Email.ValueString())
	}
	searchUsername := ""
	if hasUsername {
		searchUsername = strings.ToLower(data.Username.ValueString())
	}

	// Fetch users (Seerr API doesn't support server-side filtering, so we fetch and filter client-side)
	// We'll fetch up to 1000 users. If a server has more, this might need pagination logic, but 1000 is a safe upper bound for a typical homelab.
	res, err := d.client.Request(ctx, "GET", "/api/v1/user?take=1000", "", nil)
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

	var matchedUser map[string]any
	matchCount := 0

	for _, u := range parsedResponse.Results {
		if searchEmail != "" {
			if e, ok := u["email"].(string); ok && strings.ToLower(e) == searchEmail {
				matchedUser = u
				matchCount++
			}
		} else if searchUsername != "" {
			if un, ok := u["username"].(string); ok && strings.ToLower(un) == searchUsername {
				matchedUser = u
				matchCount++
			}
		}
	}

	if matchCount == 0 {
		resp.Diagnostics.AddError("User Not Found", "No user found matching the provided criteria.")
		return
	}
	if matchCount > 1 {
		resp.Diagnostics.AddError("Multiple Users Found", "Multiple users found matching the provided criteria. Ensure criteria is unique.")
		return
	}

	// Populate Data Source
	idRaw := matchedUser["id"]
	switch v := idRaw.(type) {
	case float64:
		data.ID = types.StringValue(fmt.Sprintf("%.0f", v))
	case string:
		data.ID = types.StringValue(v)
	default:
		resp.Diagnostics.AddError("Parse Error", "Could not parse user ID.")
		return
	}

	if e, ok := matchedUser["email"].(string); ok {
		data.Email = types.StringValue(strings.ToLower(e))
	}
	if un, ok := matchedUser["username"].(string); ok {
		data.Username = types.StringValue(strings.ToLower(un))
	}
	if p, ok := matchedUser["permissions"].(float64); ok {
		data.Permissions = types.Int64Value(int64(p))
	} else if pStr, ok := matchedUser["permissions"].(string); ok {
		if pInt, err := strconv.ParseInt(pStr, 10, 64); err == nil {
			data.Permissions = types.Int64Value(pInt)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
