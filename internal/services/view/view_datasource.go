package view

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

func NewViewDataSource() datasource.DataSource {
	return &providerConfig{}
}

type providerConfig struct {
	ConnectionCache *sql.ConnectionCache
}

func (d *providerConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_view"
}

func (d *providerConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Database view.",
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
				Description: "Name of the view",
			},
			"object_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Id of the view object in the database",
			},
			"schema": schema.StringAttribute{
				Required:    true,
				Description: "Schema where the view resides.",
			},
			"schemabinding": schema.BoolAttribute{
				Computed: true,
				Description: `If set to true, prevents the referenced objects to be changed in any way that would break 
				the functionality of the function.`,
			},
			"check_option": schema.BoolAttribute{
				Computed:    true,
				Description: `If true, all data modification statements in the view have to match the select statements. [(official docs)](https://learn.microsoft.com/en-us/sql/t-sql/statements/create-view-transact-sql?view=sql-server-ver16#check-option)`,
			},
			"definition": schema.StringAttribute{
				Computed:    true,
				Description: "Definition of the view.",
			},
		},
	}
}

func (r *providerConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state ViewDataSourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.Config.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false, true)

	if logging.HasError(ctx) {
		return
	}

	view := sql.GetViewFromNameAndSchema(ctx, connection, state.Name.ValueString(), state.Schema.ValueString(), true)

	if logging.HasError(ctx) {
		return
	}

	state.ObjectId = types.Int64Value(view.ObjectId)
	state.Id = types.StringValue(view.Id)
	state.Schemabinding = types.BoolValue(view.Schemabinding)
	state.CheckOption = types.BoolValue(view.CheckOption)
	state.Definition = types.StringValue(view.Definition)

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
