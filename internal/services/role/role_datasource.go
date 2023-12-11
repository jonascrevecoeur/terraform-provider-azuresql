package role

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

func NewRoleDataSource() datasource.DataSource {
	return &providerConfig{}
}

type providerConfig struct {
	ConnectionCache *sql.ConnectionCache
}

func (d *providerConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (d *providerConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database or server role.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the database where the role exists. database or server should be specified.",
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("server"),
					}...),
				},
			},
			"server": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the server where the user should be created. database or server should be specified.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the user",
			},
			"principal_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Principal ID of the user in the database.",
			},
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "Role or user owning the role",
			},
		},
	}
}

func (r *providerConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state RoleDataSourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.Config.Get(ctx, &state)...,
	)

	server := state.Server.ValueString()
	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database, true)
	name := state.Name.ValueString()

	if logging.HasError(ctx) {
		return
	}

	role := sql.GetRoleFromName(ctx, connection, name, true)

	if logging.HasError(ctx) {
		return
	}

	state.PrincipalId = types.Int64Value(role.PrincipalId)
	state.Owner = types.StringValue(role.Owner)
	state.Id = types.StringValue(role.Id)

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
