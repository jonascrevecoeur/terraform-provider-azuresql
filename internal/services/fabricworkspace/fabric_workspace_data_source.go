package fabricworkspace

import (
	"context"
	"fmt"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &providerConfig{}
	_ datasource.DataSourceWithConfigure = &providerConfig{}
)

func NewFabricWorkspaceDataSource() datasource.DataSource {
	return &providerConfig{}
}

type providerConfig struct {
	ConnectionCache *sql.ConnectionCache
}

func (d *providerConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_fabricworkspace"
}

func (d *providerConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Defines a connection to the fabric workspace. " +
			"Creating the data source does not yet open/test the connection. " +
			"Opening the connection happens when it is used for reading/updating another azuresql resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "ConnectionId of the fabricworkspace. " +
					"The connectionId is passed to other azuresql resources to indicate that they should use this fabricworkspace connection.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Fabric workspace. This is the value in the url preceeding `-ondemand.sql.azuresynapse.net				`",
			},
		},
	}
}

func (d *providerConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state fabricworkspaceDataSourceModel

	resp.Diagnostics.Append(
		req.Config.Get(ctx, &state)...,
	)

	var workspace string

	workspace = state.Name.ValueString()

	state.ConnectionId = types.StringValue(fmt.Sprintf("fabric::%s:1443", workspace))

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
