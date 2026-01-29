package synapseserver

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

func NewSynapseServerDataSource() datasource.DataSource {
	return &providerConfig{}
}

type providerConfig struct {
	ConnectionCache *sql.ConnectionCache
}

func (d *providerConfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_synapseserver"
}

func (d *providerConfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Defines a connection to the synapse server. " +
			"Creating the data source does not yet open/test the connection. " +
			"Opening the connection happens when it is used for reading/updating another azuresql resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				Description: "ConnectionId of the synapseserver. " +
					"The connectionId is passed to other azuresql resources to indicate that they should use this synapseserver connection.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the Synapse server. This is the value in the url preceeding `-ondemand.sql.azuresynapse.net				`",
			},
			"port": schema.Int64Attribute{
				Optional:    true,
				Description: "Port through which to connect to the synapse server (default 1433)",
			},
			"serverless": schema.BoolAttribute{
				Optional:    true,
				Description: "Use the serverless compute (default true)",
			},
		},
	}
}

func (d *providerConfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state synapseserverDataSourceModel

	resp.Diagnostics.Append(
		req.Config.Get(ctx, &state)...,
	)

	var server string
	var port int64
	var serverless bool

	server = state.Name.ValueString()

	if state.Port.IsNull() {
		port = 1433
		state.Port = types.Int64Value(1433)
	} else {
		port = state.Port.ValueInt64()
	}

	if state.Serverless.IsNull() {
		serverless = true
		state.Serverless = types.BoolValue(true)
	} else {
		serverless = state.Serverless.ValueBool()
	}

	if serverless {
		state.ConnectionId = types.StringValue(fmt.Sprintf("synapse::%s:%d", server, port))
	} else {
		state.ConnectionId = types.StringValue(fmt.Sprintf("synapsededicated::%s:%d", server, port))
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
