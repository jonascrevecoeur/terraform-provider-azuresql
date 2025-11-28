package procedure

import (
	"context"
	"fmt"
	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &providerConfig{}
	_ datasource.DataSourceWithConfigure = &providerConfig{}
)

func NewProcedureDataSource() datasource.DataSource {
	return &providerConfig{}
}

type providerConfig struct {
	ConnectionCache *sql.ConnectionCache
}

func (d *providerConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_procedure"
}

func (d *providerConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database or server user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "Id of the database where the user should be created. database or server should be specified.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the procedure",
			},
			"object_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Id of the procedure object in the database",
			},
			"schema": schema.StringAttribute{
				Required:    true,
				Description: "Schema where the procedure resides.",
			},
			"raw": schema.StringAttribute{
				Computed:    true,
				Description: "Raw definition of the procedure.",
			},
		},
	}
}

func (r *providerConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state ProcedureDataSourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.Config.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false, true)

	if logging.HasError(ctx) {
		return
	}

	procedure := sql.GetProcedureFromNameAndSchema(ctx, connection, state.Name.ValueString(), state.Schema.ValueString(), true)

	if logging.HasError(ctx) || procedure.Id == "" {
		return
	}

	state.Id = types.StringValue(procedure.Id)
	state.ObjectId = types.Int64Value(procedure.ObjectId)
	state.Raw = types.StringValue(procedure.Raw)

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
