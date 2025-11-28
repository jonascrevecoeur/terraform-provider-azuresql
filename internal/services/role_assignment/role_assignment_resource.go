package role_assignment

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

var (
	_ resource.Resource                = &RoleAssignmentResource{}
	_ resource.ResourceWithConfigure   = &RoleAssignmentResource{}
	_ resource.ResourceWithImportState = &RoleAssignmentResource{}
)

func NewRoleAssignmentResource() resource.Resource {
	return &RoleAssignmentResource{}
}

type RoleAssignmentResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *RoleAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_role_assignment"
}

func (r *RoleAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database or server role assignment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the database where the roleAssignment should be created. database or server should be specified.",
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
				Description: "Id of the server where the roleAssignment should be created. database or server should be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Azuresql resource id of the role to be used",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal": schema.StringAttribute{
				Required:    true,
				Description: "Azuresql resource id identifying the principal (user, role) who obtains the role",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RoleAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan RoleAssignmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := plan.Server.ValueString()
	database := plan.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database, true)

	if logging.HasError(ctx) {
		return
	}

	roleAssignment := sql.CreateRoleAssignment(ctx, connection, plan.Role.ValueString(), plan.Principal.ValueString())

	if logging.HasError(ctx) {
		if roleAssignment.Id != "" {
			logging.AddError(
				ctx,
				"Role assignment already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_role_assignment.<name> %s", roleAssignment.Id))
		}
		return
	}

	plan.Id = types.StringValue(roleAssignment.Id)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *RoleAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state RoleAssignmentResourceModel

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

	var roleAssignment = sql.GetRoleAssignmentFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) {
		return
	}

	if roleAssignment.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Principal = types.StringValue(roleAssignment.Principal)
	state.Role = types.StringValue(roleAssignment.Role)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *RoleAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *RoleAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	logging.AddError(ctx, "Update not implemented", "Update should never be called for roleAssignment resources as any change requires a delete and recreate.")
}

func (r *RoleAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state RoleAssignmentResourceModel

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

	sql.DropRoleAssignment(ctx, connection, state.Id.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping roleAssignment failed", fmt.Sprintf("Dropping roleAssignment %s failed", state.Id.ValueString()))
	}
}

func (r *RoleAssignmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	roleAssignment := sql.ParseRoleAssignmentId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := sql.ParseConnectionId(ctx, roleAssignment.Connection)

	if logging.HasError(ctx) {
		return
	}

	connection = r.ConnectionCache.Connect(ctx, connection.ConnectionId, connection.IsServerConnection, true)

	if logging.HasError(ctx) {
		return
	}

	roleAssignment = sql.GetRoleAssignmentFromId(ctx, connection, req.ID, true)

	if logging.HasError(ctx) {
		return
	}

	state := RoleAssignmentResourceModel{
		Id:        types.StringValue(roleAssignment.Id),
		Role:      types.StringValue(roleAssignment.Role),
		Principal: types.StringValue(roleAssignment.Principal),
	}

	if connection.IsServerConnection {
		state.Server = types.StringValue(roleAssignment.Connection)
	} else {
		state.Database = types.StringValue(roleAssignment.Connection)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
