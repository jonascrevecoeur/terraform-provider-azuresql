package permission

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
	_ resource.Resource                = &PermissionResource{}
	_ resource.ResourceWithConfigure   = &PermissionResource{}
	_ resource.ResourceWithImportState = &PermissionResource{}
)

func NewPermissionResource() resource.Resource {
	return &PermissionResource{}
}

type PermissionResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *PermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_permission"
}

func (r *PermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database or server permission.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the database where the permission should be created. database or server should be specified.",
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
				Description: "Id of the server where the permission should be created. database or server should be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"scope": schema.StringAttribute{
				Required:    true,
				Description: "Azuresql resource id determining the scope of the permission (table, view, schema, database, server)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal": schema.StringAttribute{
				Required:    true,
				Description: "Azuresql resource id having the permission (user, role)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"permission": schema.StringAttribute{
				Required:    true,
				Description: "Permission to be granted.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *PermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan PermissionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	server := plan.Server.ValueString()
	database := plan.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database)

	if logging.HasError(ctx) {
		return
	}

	permission := sql.CreatePermission(ctx, connection, plan.Scope.ValueString(), plan.Principal.ValueString(), plan.Permission.ValueString())

	if logging.HasError(ctx) {
		return
	}

	plan.Id = types.StringValue(permission.Id)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *PermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state PermissionResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	server := state.Server.ValueString()
	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database)

	if logging.HasError(ctx) {
		return
	}

	var permission = sql.GetPermissionFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) || permission.Id == "" {
		if permission.Id != "" {
			logging.AddError(
				ctx,
				"Permission already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_permission.<name> %s", permission.Id))
		}
		return
	}

	state.Principal = types.StringValue(permission.Principal)
	state.Scope = types.StringValue(permission.Scope)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *PermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *PermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	logging.AddError(ctx, "Update not implemented", "Update should never be called for permission resources as any change requires a delete and recreate.")
}

func (r *PermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state PermissionResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	server := state.Server.ValueString()
	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database)

	if logging.HasError(ctx) {
		return
	}

	sql.DropPermission(ctx, connection, state.Scope.ValueString(), state.Principal.ValueString(), state.Permission.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping permission failed", fmt.Sprintf("Dropping permission %s failed", state.Permission.ValueString()))
	}
}

func (r *PermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	permission := sql.ParsePermissionId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := sql.ParseConnectionId(ctx, permission.Connection)

	if logging.HasError(ctx) {
		return
	}

	connection = r.ConnectionCache.Connect(ctx, connection.ConnectionId, connection.IsServerConnection)

	if logging.HasError(ctx) {
		return
	}

	permission = sql.GetPermissionFromId(ctx, connection, req.ID, true)

	if logging.HasError(ctx) {
		return
	}

	state := PermissionResourceModel{
		Id:         types.StringValue(permission.Id),
		Scope:      types.StringValue(permission.Scope),
		Principal:  types.StringValue(permission.Principal),
		Permission: types.StringValue(permission.Permission),
	}

	if connection.IsServerConnection {
		state.Server = types.StringValue(permission.Connection)
	} else {
		state.Database = types.StringValue(permission.Connection)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}
