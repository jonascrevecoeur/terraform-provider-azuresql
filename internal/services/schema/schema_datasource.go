package schema

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

func NewSchemaDataSource() datasource.DataSource {
	return &providerConfig{}
}

type providerConfig struct {
	ConnectionCache *sql.ConnectionCache
}

func (d *providerConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schema"
}

func (d *providerConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database schema.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the database where the schema exists.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the schema.",
			},
			"schema_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Schema ID of the schema in the database.",
			},
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "Principal owning the schema.",
			},
		},
	}
}

func (r *providerConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state SchemaDataSourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.Config.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	schema := sql.GetSchemaFromName(ctx, connection, state.Name.ValueString(), true)

	if logging.HasError(ctx) {
		return
	}

	state.SchemaId = types.Int64Value(schema.SchemaId)
	state.Owner = types.StringValue(schema.Owner)
	state.Id = types.StringValue(schema.Id)

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
