package external_data_source

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

func NewExternalDataSourceDataSource() datasource.DataSource {
	return &providerConfig{}
}

type providerConfig struct {
	ConnectionCache *sql.ConnectionCache
}

func (d *providerConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_data_source"
}

func (d *providerConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read an external data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "Id of the database where the external data source should be created.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the external data source",
			},
			"data_source_id": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the external data source in the database.",
			},
			"location": schema.StringAttribute{
				Computed:    true,
				Description: "Location of the external data source.",
			},
			"credential": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the azuresql_database_scoped_credential used to access the external data source.",
			},
		},
	}
}

func (r *providerConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state ExternalDataSourceDataSourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.Config.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	externalDataSource := sql.GetExternalDataSourceFromName(ctx, connection, state.Name.ValueString(), true)

	if logging.HasError(ctx) {
		return
	}

	state.DataSourceId = types.Int64Value(externalDataSource.DataSourceId)
	state.Id = types.StringValue(externalDataSource.Id)
	state.Location = types.StringValue(externalDataSource.Location)
	if externalDataSource.Credential != "" {
		state.Credential = types.StringValue(externalDataSource.Credential)
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
