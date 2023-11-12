package external_data_source

import (
	"context"
	"fmt"

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
	_ resource.Resource                = &ExternalDataSourceResource{}
	_ resource.ResourceWithConfigure   = &ExternalDataSourceResource{}
	_ resource.ResourceWithImportState = &ExternalDataSourceResource{}
)

func NewExternalDataSourceResource() resource.Resource {
	return &ExternalDataSourceResource{}
}

type ExternalDataSourceResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *ExternalDataSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_external_data_source"
}

func (r *ExternalDataSourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Register an external data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "Id of the database where the external data source should be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the external data source",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data_source_id": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the external data source in the database.",
			},
			"location": schema.StringAttribute{
				Required:    true,
				Description: "Location of the external data source.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"credential": schema.StringAttribute{
				Optional:    true,
				Description: "ID of the azuresql_database_scoped_credential used to access the external data source.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ExternalDataSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan ExternalDataSourceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	database := plan.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false, true)

	if logging.HasError(ctx) {
		return
	}

	externalDataSource := sql.CreateExternalDataSource(ctx, connection, name, plan.Location.ValueString(), plan.Credential.ValueString())

	if logging.HasError(ctx) {
		if externalDataSource.Id != "" {
			logging.AddError(
				ctx,
				"ExternalDataSource already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_externalDataSource.<name> %s", externalDataSource.Id))
		}
		return
	}

	plan.Id = types.StringValue(externalDataSource.Id)
	plan.DataSourceId = types.Int64Value(externalDataSource.DataSourceId)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ExternalDataSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	var state ExternalDataSourceResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false, false)

	if logging.HasError(ctx) {
		return
	}

	if connection.ConnectionResourceStatus == sql.ConnectionResourceStatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	externalDataSource := sql.GetExternalDataSourceFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) {
		return
	}

	if externalDataSource.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.DataSourceId = types.Int64Value(externalDataSource.DataSourceId)
	state.Name = types.StringValue(externalDataSource.Name)
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

func (r *ExternalDataSourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *ExternalDataSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	logging.AddError(ctx, "Update not implemented", "Update should never be called for external data source resources.")
}

func (r *ExternalDataSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state ExternalDataSourceResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false, false)

	if connection.ConnectionResourceStatus == sql.ConnectionResourceStatusNotFound {
		return
	}

	if logging.HasError(ctx) {
		return
	}

	sql.DropExternalDataSource(ctx, connection, state.DataSourceId.ValueInt64())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping external data source failed", fmt.Sprintf("Dropping external data source %s failed", state.Name.ValueString()))
	}
}

func (r *ExternalDataSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	externalDataSource := sql.ParseExternalDataSourceId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := sql.ParseConnectionId(ctx, externalDataSource.Connection)

	if logging.HasError(ctx) {
		return
	}

	connection = r.ConnectionCache.Connect(ctx, connection.ConnectionId, false, true)

	if logging.HasError(ctx) {
		return
	}

	externalDataSource = sql.GetExternalDataSourceFromId(ctx, connection, req.ID, true)

	if logging.HasError(ctx) {
		return
	}

	state := ExternalDataSourceResourceModel{
		Id:           types.StringValue(externalDataSource.Id),
		Name:         types.StringValue(externalDataSource.Name),
		Database:     types.StringValue(externalDataSource.Connection),
		DataSourceId: types.Int64Value(externalDataSource.DataSourceId),
		Location:     types.StringValue(externalDataSource.Location),
	}

	if externalDataSource.Credential != "" {
		state.Credential = types.StringValue(externalDataSource.Credential)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
