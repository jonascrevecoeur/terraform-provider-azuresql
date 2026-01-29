package database

import (
	"context"
	"fmt"
	"strings"

	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &DatabaseResource{}
	_ resource.ResourceWithConfigure   = &DatabaseResource{}
	_ resource.ResourceWithImportState = &DatabaseResource{}
)

func NewDatabaseResource() resource.Resource {
	return &DatabaseResource{}
}

type DatabaseResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *DatabaseResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database"
}

func (r *DatabaseResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"server": schema.StringAttribute{
				Required:    true,
				Description: "Id of the server resource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the database within the server",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DatabaseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan DatabaseResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	server := plan.Server.ValueString()
	connection := r.ConnectionCache.Connect(ctx, server, true, true)

	if logging.HasError(ctx) {
		return
	}

	if connection.Provider == "sqlserver" {
		logging.AddError(ctx, "Invalid config",
			"`azuresql_database` does not support creating databases in sqlserver. Use the `azurerm_mssql_database` resource from the `azurerm` provider instead.")
		return
	}

	if connection.Provider == "synapsededicated" {
		logging.AddError(ctx, "Invalid config",
			"`azuresql_database` does not support creating databases in Synapse dedicated servers.")
		return
	}

	if connection.Provider == "fabric" {
		logging.AddError(ctx, "Invalid config",
			"`azuresql_database` does not support creating Fabric databases.")
		return
	}

	database := sql.CreateDatabase(ctx, connection, name)

	if logging.HasError(ctx) {
		if database.Id != "" {
			logging.AddError(
				ctx,
				"Database already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_database.<name> %s", database.Id))
		}
		return
	}

	plan.ConnectionId = types.StringValue(database.Id)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DatabaseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state DatabaseResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	connection := sql.ParseConnectionId(ctx, state.ConnectionId.ValueString())
	if logging.HasError(ctx) {
		return
	}

	status := r.ConnectionCache.DatabaseExists(ctx, connection)

	if logging.HasError(ctx) {
		return
	}

	if status == sql.ConnectionResourceStatusNotFound {
		resp.State.RemoveResource(ctx)
	}

	return
}

func (r *DatabaseResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

	if req.ProviderData == nil {
		return
	}

	cache, ok := req.ProviderData.(*sql.ConnectionCache)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *sql.Server, got: %T.", req.ProviderData),
		)

		return
	}

	r.ConnectionCache = cache
}

func (r *DatabaseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	logging.AddError(ctx, "Update not implemented", "Update should never be called for database resources as any change requires a delete and recreate.")
}

func (r *DatabaseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state DatabaseResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	connection := r.ConnectionCache.Connect(ctx, state.Server.ValueString(), true, false)

	if logging.HasError(ctx) {
		return
	}

	if connection.ConnectionResourceStatus == sql.ConnectionResourceStatusNotFound {
		return
	}

	sql.DropDatabase(ctx, connection, state.Name.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping database failed", fmt.Sprintf("Dropping database %s failed", state.Name.ValueString()))
	}
}

func (r *DatabaseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	connection := sql.ParseConnectionId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	if connection.Provider == "sqlserver" {
		logging.AddError(ctx, "Invalid config",
			"`azuresql_database` does not support creation & deletion of databases in sqlserver. Use the `azurerm_mssql_database` resource from the `azurerm` provider instead.")
		return
	}

	status := r.ConnectionCache.DatabaseExists(ctx, connection)

	if logging.HasError(ctx) {
		return
	}

	if status == sql.ConnectionResourceStatusNotFound {
		logging.AddError(ctx, "Database not found", fmt.Sprintf("Database with id %s not found", connection.ConnectionId))
		return
	}

	state := DatabaseResourceModel{
		ConnectionId: types.StringValue(connection.ConnectionId),
		Server:       types.StringValue(strings.TrimSuffix(connection.ConnectionId, ":"+connection.Database)),
		Name:         types.StringValue(connection.Database),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
