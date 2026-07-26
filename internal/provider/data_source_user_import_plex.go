package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserImportPlexDataSource{}

type UserImportPlexDataSource struct {
	client *APIClient
}

type UserImportPlexDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ImportedUsers types.List   `tfsdk:"imported_users"`
	ImportedCount types.Int64  `tfsdk:"imported_count"`
}

func NewUserImportPlexDataSource() datasource.DataSource {
	return &UserImportPlexDataSource{}
}

func (d *UserImportPlexDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_import_plex"
}

func (d *UserImportPlexDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Read users imported from a connected Plex server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier.",
				Computed:            true,
			},
			"imported_users": schema.ListNestedAttribute{
				MarkdownDescription: "List of users imported from Plex.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The numeric Seerr user ID.",
							Computed:            true,
						},
						"username": schema.StringAttribute{
							MarkdownDescription: "The user's display name.",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "The user's email address.",
							Computed:            true,
						},
						"plex_id": schema.StringAttribute{
							MarkdownDescription: "The user's Plex ID.",
							Computed:            true,
						},
					},
				},
			},
			"imported_count": schema.Int64Attribute{
				MarkdownDescription: "Number of Plex users imported.",
				Computed:            true,
			},
		},
	}
}

func (d *UserImportPlexDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserImportPlexDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserImportPlexDataSourceModel

	results, err := fetchAllPaginatedResults(ctx, d.client, "/api/v1/user", defaultPaginationPageSize)
	if err != nil {
		resp.Diagnostics.AddError("Read Failed", err.Error())
		return
	}

	var imported []ImportedPlexUserModel
	for _, u := range results {
		plexID, _ := u["plexId"].(string)
		if plexID == "" {
			continue
		}

		user := ImportedPlexUserModel{
			PlexID: types.StringValue(plexID),
		}

		switch v := u["id"].(type) {
		case float64:
			user.ID = types.StringValue(fmt.Sprintf("%.0f", v))
		case string:
			user.ID = types.StringValue(v)
		}

		if e, ok := u["email"].(string); ok {
			user.Email = types.StringValue(e)
		}
		if un, ok := u["username"].(string); ok {
			user.Username = types.StringValue(un)
		}

		imported = append(imported, user)
	}

	listVal, diags := buildImportedPlexUserListValue(ctx, imported)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue("plex_users")
	data.ImportedUsers = listVal
	data.ImportedCount = types.Int64Value(int64(len(imported)))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
