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
    _ datasource.DataSource              = &providerConfigDedicated{}
    _ datasource.DataSourceWithConfigure = &providerConfigDedicated{}
)

func NewSynapseServerDedicatedDataSource() datasource.DataSource {
    return &providerConfigDedicated{}
}

type providerConfigDedicated struct {
    ConnectionCache *sql.ConnectionCache
}

func (d *providerConfigDedicated) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_synapseserver_dedicated"
}

func (d *providerConfigDedicated) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Defines a connection to an Azure Synapse Dedicated SQL endpoint (SQL pools). " +
            "Creating the data source does not yet open/test the connection. " +
            "Opening the connection happens when it is used for reading/updating another azuresql resource.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:    true,
                Description: "ConnectionId of the Synapse Dedicated endpoint. Pass this id to other azuresql resources to use this connection.",
            },
            "name": schema.StringAttribute{
                Required:    true,
                Description: "Name of the Synapse workspace (the hostname prefix of '<workspace>.sql.azuresynapse.net').",
            },
            "port": schema.Int64Attribute{
                Optional:    true,
                Description: "Port to connect to the Synapse Dedicated endpoint (default 1433).",
            },
        },
    }
}

func (d *providerConfigDedicated) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var state synapseserverDataSourceModel

    resp.Diagnostics.Append(
        req.Config.Get(ctx, &state)...,
    )

    var server string
    var port int64

    server = state.Name.ValueString()

    if state.Port.IsNull() {
        port = 1433
        state.Port = types.Int64Value(1433)
    } else {
        port = state.Port.ValueInt64()
    }

    state.ConnectionId = types.StringValue(fmt.Sprintf("synapsededicated::%s:%d", server, port))

    diags := resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if resp.Diagnostics.HasError() {
        return
    }
}

func (d *providerConfigDedicated) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

