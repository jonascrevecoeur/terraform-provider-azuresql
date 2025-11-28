package login

import (
	"context"
	"fmt"
	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ datasource.DataSource              = &providerConfig{}
	_ datasource.DataSourceWithConfigure = &providerConfig{}
)

func NewSQLLoginDataSource() datasource.DataSource {
	return &providerConfig{}
}

type providerConfig struct {
	ConnectionCache *sql.ConnectionCache
}

func (d *providerConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login"
}

func (d *providerConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Logins are used to authenticate SQL users.
		Logins can only be created on the server level, but can be used to create database users.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"server": schema.StringAttribute{
				Required:    true,
				Description: "Id of the server on which this login exists.",
			},
			"sid": schema.StringAttribute{
				Computed:    true,
				Description: "sid assocatied to the login on the server",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Login name",
			},
		},
	}
}

func (r *providerConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state SQLLoginDataSourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.Config.Get(ctx, &state)...,
	)

	connectionId := state.Server.ValueString()
	tflog.Info(ctx, fmt.Sprintf("ConnectionId: %s", connectionId))
	connection := r.ConnectionCache.Connect(ctx, connectionId, true, true)
	name := state.Name.ValueString()

	if logging.HasError(ctx) {
		return
	}

	login := sql.GetLoginFromName(ctx, connection, name)

	if logging.HasError(ctx) {
		return
	}

	if login.Id == "" {
		logging.AddError(ctx, "Datasource not found", fmt.Sprintf("Login %s not found on server %s", name, connectionId))
		return
	}

	state.Sid = types.StringValue(login.Sid)
	state.Id = types.StringValue(login.Id)

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
