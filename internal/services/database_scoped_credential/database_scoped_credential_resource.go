package database_scoped_credential

import (
	"context"
	"fmt"

	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &DatabaseScopedCredentialResource{}
	_ resource.ResourceWithConfigure = &DatabaseScopedCredentialResource{}
)

func NewDatabaseScopedCredentialResource() resource.Resource {
	return &DatabaseScopedCredentialResource{}
}

type DatabaseScopedCredentialResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *DatabaseScopedCredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_database_scoped_credential"
}

func (r *DatabaseScopedCredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database scoped credential.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "Id of the database where the database scoped credential should be created.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name identifier of the database scoped credential in the database.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"identity": schema.StringAttribute{
				Required:    true,
				Description: "Identify of the datbase scoped credential.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"secret": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Secret for the database scoped credential. Leave blank to use no secret.",
				Sensitive:   true,
			},
			"credential_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Id of the database scoped credential in the database",
			},
		},
	}
}

func (r *DatabaseScopedCredentialResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan DatabaseScopedCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	database := plan.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false, true)

	if connection.Provider == "fabric" {
		logging.AddError(ctx, "invalid config", "Credentials are not supported in Fabric")
		return
	}

	if logging.HasError(ctx) {
		return
	}

	databaseScopedCredential := sql.CreateDatabaseScopedCredential(
		ctx,
		connection,
		plan.Name.ValueString(),
		plan.Identity.ValueString(),
		plan.Secret.ValueString(),
	)

	if logging.HasError(ctx) {
		if databaseScopedCredential.Id != "" {
			logging.AddError(
				ctx,
				"Database scoped credential already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_database_scoped_credential.<name> %s", databaseScopedCredential.Id))
		}
		return
	}

	plan.Id = types.StringValue(databaseScopedCredential.Id)
	plan.CredentialId = types.Int64Value(databaseScopedCredential.CredentialId)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DatabaseScopedCredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state DatabaseScopedCredentialResourceModel

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

	databaseScopedCredential := sql.GetDatabaseScopedCredentialFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) {
		return
	}

	if databaseScopedCredential.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(databaseScopedCredential.Name)
	state.Identity = types.StringValue(databaseScopedCredential.Identity)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *DatabaseScopedCredentialResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *DatabaseScopedCredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan DatabaseScopedCredentialResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	database := plan.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false, true)

	if logging.HasError(ctx) {
		return
	}

	databaseScopedCredential := sql.AlterDatabaseScopedCredential(
		ctx,
		connection,
		plan.Name.ValueString(),
		plan.Identity.ValueString(),
		plan.Secret.ValueString(),
	)

	if logging.HasError(ctx) {
		return
	}

	plan.Id = types.StringValue(databaseScopedCredential.Id)
	plan.CredentialId = types.Int64Value(databaseScopedCredential.CredentialId)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DatabaseScopedCredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state DatabaseScopedCredentialResourceModel

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
		return
	}

	sql.DropDatabaseScopedCredential(ctx, connection, state.Id.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping database scoped credential failed", fmt.Sprintf("Dropping database scoped credential failed"))
	}
}

func (r *DatabaseScopedCredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	databaseScopedCredential := sql.ParseDatabaseScopedCredentialId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, databaseScopedCredential.Connection, false, true)

	if logging.HasError(ctx) {
		return
	}

	databaseScopedCredential = sql.GetDatabaseScopedCredentialFromId(ctx, connection, req.ID, true)

	if logging.HasError(ctx) {
		return
	}

	state := DatabaseScopedCredentialResourceModel{
		Id:           types.StringValue(databaseScopedCredential.Id),
		Database:     types.StringValue(databaseScopedCredential.Connection),
		Name:         types.StringValue(databaseScopedCredential.Name),
		Identity:     types.StringValue(databaseScopedCredential.Identity),
		CredentialId: types.Int64Value(databaseScopedCredential.CredentialId),
		Secret:       types.StringValue(""),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
