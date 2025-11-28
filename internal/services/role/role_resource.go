package role

import (
	"context"
	"fmt"

	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &RoleResource{}
	_ resource.ResourceWithConfigure   = &RoleResource{}
	_ resource.ResourceWithImportState = &RoleResource{}
)

func NewRoleResource() resource.Resource {
	return &RoleResource{}
}

type RoleResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *RoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role"
}

func (r *RoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database or server user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the database where the user should be created. database or server should be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("server"),
					}...),
				},
			},
			"server": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the server where the user should be created. database or server should be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the user",
			},
			"principal_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Principal ID of the user in the database.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Role or user owning the role.",
			},
		},
	}
}

func (r RoleResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	// no modification required on create or delete
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state RoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := state.Server.ValueString()
	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database, false)

	// in Synapse serverless alter authorization cannot be used
	// -> a replace is required when owner changes
	if connection.Provider == "synapse" {
		resp.RequiresReplace.Append(path.Root("owner"))
	}
}

func (r *RoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan RoleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	server := plan.Server.ValueString()
	database := plan.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database, true)

	if logging.HasError(ctx) {
		return
	}

	if connection.IsServerConnection && connection.Provider == "synapse" {
		logging.AddError(ctx, "Invalid config",
			"In Synapse users cannot be created at server level. Try creating a database user instead.")
		return
	}

	owner := plan.Owner.ValueString()

	role := sql.CreateRole(ctx, connection, name, owner)

	if logging.HasError(ctx) {
		if role.Id != "" {
			logging.AddError(
				ctx,
				"Role already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_role.<name> %s", role.Id))
		}
		return
	}

	plan.Id = types.StringValue(role.Id)
	plan.PrincipalId = types.Int64Value(role.PrincipalId)
	plan.Owner = types.StringValue(role.Owner)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *RoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	var state RoleResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	server := state.Server.ValueString()
	database := state.Database.ValueString()

	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database, false)

	if logging.HasError(ctx) {
		return
	}

	if connection.ConnectionResourceStatus == sql.ConnectionResourceStatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	role := sql.GetRoleFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) {
		return
	}

	if role.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.PrincipalId = types.Int64Value(role.PrincipalId)
	state.Name = types.StringValue(role.Name)
	state.Owner = types.StringValue(role.Owner)
	state.Id = types.StringValue(role.Id)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *RoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *RoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state RoleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan RoleResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// changes in these values would have triggered a replacement
	// so they are identical for state/plan
	server := state.Server.ValueString()
	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database, true)
	id := state.Id.ValueString()

	// update name
	if !state.Name.Equal(plan.Name) {
		sql.UpdateRoleName(ctx, connection, id, plan.Name.ValueString())

		if logging.HasError(ctx) {
			return
		}
		state.Name = plan.Name
	}

	// update owner
	if !state.Owner.Equal(plan.Owner) && plan.Owner.ValueString() != "" {

		sql.UpdateRoleOwner(ctx, connection, id, plan.Owner.ValueString())

		if logging.HasError(ctx) {
			return
		}
		state.Owner = plan.Owner
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *RoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state RoleResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	server := state.Server.ValueString()
	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database, false)

	if logging.HasError(ctx) {
		return
	}

	if connection.ConnectionResourceStatus == sql.ConnectionResourceStatusNotFound {
		return
	}

	sql.DropRole(ctx, connection, state.PrincipalId.ValueInt64())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping role failed", fmt.Sprintf("Dropping role %s failed", state.Name.ValueString()))
	}
}

func (r *RoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	role := sql.ParseRoleId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := sql.ParseConnectionId(ctx, role.Connection)

	if logging.HasError(ctx) {
		return
	}

	connection = r.ConnectionCache.Connect(ctx, connection.ConnectionId, connection.IsServerConnection, true)

	if logging.HasError(ctx) {
		return
	}

	role = sql.GetRoleFromId(ctx, connection, req.ID, true)

	if logging.HasError(ctx) {
		return
	}

	state := RoleResourceModel{
		Id:          types.StringValue(role.Id),
		Name:        types.StringValue(role.Name),
		PrincipalId: types.Int64Value(role.PrincipalId),
		Owner:       types.StringValue(role.Owner),
	}

	if connection.IsServerConnection {
		state.Server = types.StringValue(role.Connection)
	} else {
		state.Database = types.StringValue(role.Connection)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
