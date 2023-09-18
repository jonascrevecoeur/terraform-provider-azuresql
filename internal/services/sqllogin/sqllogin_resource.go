package login

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
/*var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)*/

func NewSQLLoginResource() resource.Resource {
	return &SQLLoginResource{}
}

type SQLLoginResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *SQLLoginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login"
}

func (r *SQLLoginResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Register a user in the database. Implemented for `SQL Server` and `Azure SQL Database`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"server": schema.StringAttribute{
				Required:    true,
				Description: "Id of the server on which this login exists.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"sid": schema.StringAttribute{
				Computed:    true,
				Description: "sid assocatied to the login on the server",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Login name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Computed:    true,
				Description: "Auto generated password for the new login.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *SQLLoginResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan SQLLoginResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	server := plan.Server.ValueString()
	connection := r.ConnectionCache.Connect(ctx, server, true)

	if logging.HasError(ctx) {
		return
	}

	login := sql.CreateLogin(ctx, connection, name)

	if logging.HasError(ctx) {
		return
	}

	plan.Password = types.StringValue(login.Password)
	plan.Sid = types.StringValue(login.Sid)
	plan.Id = types.StringValue(login.Id)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SQLLoginResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state SQLLoginResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	connectionId := state.Server.ValueString()
	connection := r.ConnectionCache.Connect(ctx, connectionId, true)

	var login sql.Login
	if state.Sid.IsNull() && state.Name.IsNull() {
		logging.AddError(ctx, "Unable to read azuresql_login", "Cannot read login when both sid and name are unknown")
		return
	} else if state.Sid.IsNull() {
		// if both name and sid are available prefer sid
		login = sql.GetLoginFromName(ctx, connection, state.Name.ValueString())
	} else {
		login = sql.GetLoginFromSid(ctx, connection, state.Sid.ValueString())
	}

	if logging.HasError(ctx) || login.Id == "" {
		return
	}

	state.Sid = types.StringValue(login.Sid)
	state.Name = types.StringValue(login.Name)
	state.Id = types.StringValue(login.Id)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SQLLoginResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *SQLLoginResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Never triggered as a replace is always required
}

func (r *SQLLoginResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state SQLLoginResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	connection := r.ConnectionCache.Connect(ctx, state.Server.ValueString(), true)

	sql.DropLogin(ctx, connection, state.Sid.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping user failed", fmt.Sprintf("Dropping user %s failed", state.Name.ValueString()))
	}
}

func (r *SQLLoginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	/*ctx = utils.WithDiagnostics(ctx, &resp.Diagnostics)

	user := sql._user_parse_id(ctx, req.ID)

	if utils.HasError(ctx) {
		return
	}

	resp.State.SetAttribute(ctx, path.Root("connection_string"), user.ConnectionString)
	resp.State.SetAttribute(ctx, path.Root("principal_id"), user.PrincipalId)*/
}
