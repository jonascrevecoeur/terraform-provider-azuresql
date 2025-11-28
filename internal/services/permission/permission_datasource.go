package permission

import (
	"context"
	"fmt"
	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &providerConfig{}
	_ datasource.DataSourceWithConfigure = &providerConfig{}
)

func NewPermissionDataSource() datasource.DataSource {
	return &providerConfig{}
}

type providerConfig struct {
	ConnectionCache *sql.ConnectionCache
}

func (d *providerConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (d *providerConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database or server permission.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Id used for testing the provider, this id cannot be used to import resources.",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the database where the permission should be created. database or server should be specified.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("server"),
					}...),
				},
			},
			"server": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the server where the permission should be created. database or server should be specified.",
			},
			"scope": schema.StringAttribute{
				Required:    true,
				Description: "Azuresql resource id determining the scope of the permission (table, view, schema, database, server)",
			},
			"principal": schema.StringAttribute{
				Required:    true,
				Description: "Azuresql resource id having the permission (user, role)",
			},
			"permissions": schema.ListAttribute{
				Computed:    true,
				Description: "List of granted permissions.",
				ElementType: types.StringType,
			},
		},
	}
}

func (r *providerConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state PermissionDataSourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.Config.Get(ctx, &state)...,
	)

	server := state.Server.ValueString()
	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database, true)

	if logging.HasError(ctx) {
		return
	}

	var permissions = sql.GetAllPermissions(ctx, connection, state.Scope.ValueString(), state.Principal.ValueString())

	if logging.HasError(ctx) {
		return
	}

	state.Permissions = []types.String{}
	state.Id = types.StringValue(
		fmt.Sprintf("%s/permissions/scope:%s/principal:%s", connection.ConnectionId, state.Scope.ValueString(), state.Principal.ValueString()))

	for _, permission := range permissions {
		state.Permissions = append(state.Permissions, types.StringValue(permission))
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *providerConfig) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cache, ok := req.ProviderData.(*sql.ConnectionCache)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sql.ConnectionCache, got: %T.", req.ProviderData),
		)

		return
	}

	d.ConnectionCache = cache
}
