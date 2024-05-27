package user

import (
	"context"
	"fmt"

	"terraform-provider-azuresql/internal/docu"
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
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
/*var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)*/

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: fmt.Sprintf("SQL database or server user. %s", docu.Supported(true, true, true, true)),
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Description: docu.FieldDatabase("user", true, true),
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
				Description: "Id of the server where the user should be created test. `database` or server should be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the user",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"principal_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Principal ID of the user in the database.",
			},
			"authentication": schema.StringAttribute{
				Required:    true,
				Description: "The user authentication mode. Possible values are `AzureAD`, `SQLLogin` and `WithoutLogin`.",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"AzureAD", "SQLLogin", "WithoutLogin"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of the user in the database. Possible types are TODO.",
			},
			"login": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the server where the user should be created. database or server should be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r UserResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var data UserResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Login.IsNull() && data.Authentication.ValueString() != "SQLLogin" {
		logging.AddAttributeError(ctx, path.Root("login"), "Invalid attribute configuration",
			"login is only allowed when authentication equals `SQLLogin`")
		return
	}

	if data.Login.IsNull() && data.Authentication.ValueString() == "SQLLogin" {
		logging.AddAttributeError(ctx, path.Root("login"), "Invalid attribute configuration",
			"login is required when authentication equals `SQLLogin`")
		return
	}
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan UserResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	server := plan.Server.ValueString()
	database := plan.Database.ValueString()
	connection := r.ConnectionCache.Connect_server_or_database(ctx, server, database, true)
	tflog.Debug(ctx, fmt.Sprintf("Status %d", connection.ConnectionResourceStatus))
	if logging.HasError(ctx) {
		return
	}

	if connection.IsServerConnection && connection.Provider == "synapse" {
		logging.AddError(ctx, "Invalid config", "In Synapse users cannot be created at server level. Try creating a database user instead.")
		return
	}

	authentication := plan.Authentication.ValueString()
	login := plan.Login.ValueString()

	user := sql.CreateUser(ctx, connection, name, authentication, login)

	if logging.HasError(ctx) {
		if user.Id != "" {
			logging.AddError(
				ctx,
				"User assignment already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_user.<name> %s", user.Id))
		}
		return
	}

	plan.Id = types.StringValue(user.Id)
	plan.PrincipalId = types.Int64Value(user.PrincipalId)
	plan.Type = types.StringValue(user.Type)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state UserResourceModel

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

	var user sql.User
	if state.PrincipalId.IsNull() && state.Name.IsNull() {
		logging.AddError(ctx, "Unable to read azuresql_user", "Cannot read user when both id and name are unknown")
		return
	} else if state.PrincipalId.IsNull() {
		// if both name and principalId are available prefer principalId
		user = sql.GetUserFromName(ctx, connection, state.Name.ValueString())
	} else {
		user = sql.GetUserFromPrincipalId(ctx, connection, state.PrincipalId.ValueInt64())
	}

	if logging.HasError(ctx) {
		return
	}

	if user.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.PrincipalId = types.Int64Value(user.PrincipalId)
	state.Name = types.StringValue(user.Name)
	state.Type = types.StringValue(user.Type)
	state.Authentication = types.StringValue(user.Authentication)

	state.Id = types.StringValue(user.Id)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Never triggered as a replace is always required
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state UserResourceModel

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

	sql.DropUser(ctx, connection, state.PrincipalId.ValueInt64())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping user failed", fmt.Sprintf("Dropping user %s failed", state.Name.ValueString()))
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	user := sql.ParseUserId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := sql.ParseConnectionId(ctx, user.Connection)

	if logging.HasError(ctx) {
		return
	}

	connection = r.ConnectionCache.Connect(ctx, connection.ConnectionId, connection.IsServerConnection, true)

	if logging.HasError(ctx) {
		return
	}

	user = sql.GetUserFromId(ctx, connection, req.ID, true)

	if logging.HasError(ctx) {
		return
	}

	state := UserResourceModel{
		Id:             types.StringValue(user.Id),
		Name:           types.StringValue(user.Name),
		PrincipalId:    types.Int64Value(user.PrincipalId),
		Authentication: types.StringValue(user.Authentication),
		Type:           types.StringValue(user.Type),
	}

	if user.Login != "" {
		state.Login = types.StringValue(user.Login)
	}

	if connection.IsServerConnection {
		state.Server = types.StringValue(user.Connection)
	} else {
		state.Database = types.StringValue(user.Connection)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
