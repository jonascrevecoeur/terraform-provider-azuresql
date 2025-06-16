package login

import (
	"context"
	"fmt"

	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &SQLLoginResource{}
	_ resource.ResourceWithConfigure   = &SQLLoginResource{}
	_ resource.ResourceWithImportState = &SQLLoginResource{}
)

func NewSQLLoginResource() resource.Resource {
	return &SQLLoginResource{}
}

type SQLLoginResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *SQLLoginResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_login"
}

func (r *SQLLoginResource) SchemaPropertiesAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"arguments": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name": types.StringType,
					"type": types.StringType,
				},
			},
		},
		"definition": types.StringType,
		"executor":   types.StringType,
	}
}

func (r *SQLLoginResource) SchemaPasswordProperties() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"length": schema.Int32Attribute{
			Optional: true,
			Computed: true,
			Default:  int32default.StaticInt32(20),
		},
		"allowed_special_chars": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString(sql.SpecialCharSet),
		},
		"min_special_chars": schema.Int32Attribute{
			Optional: true,
			Computed: true,
			Default:  int32default.StaticInt32(3),
		},
		"min_numbers": schema.Int32Attribute{
			Optional: true,
			Computed: true,
			Default:  int32default.StaticInt32(4),
		},
		"min_uppercase": schema.Int32Attribute{
			Optional: true,
			Computed: true,
			Default:  int32default.StaticInt32(5),
		},
	}
}

func (r *SQLLoginResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: `Logins are used to authenticate SQL users.
		Logins can only be created on the server level, but can be used to create database users.`,
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
				Sensitive: true,
			},
			"password_properties": schema.SingleNestedAttribute{
				Optional: true,
				// Disabled for now - requires proper parsing of raw to props
				// Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: r.SchemaPasswordProperties(),
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
	connection := r.ConnectionCache.Connect(ctx, server, true, true)

	if connection.Provider == "fabric" {
		logging.AddError(ctx, "invalid config", "Managing logins is not supported in Fabric")
		return
	}

	if logging.HasError(ctx) {
		return
	}

	login := sql.CreateLogin(ctx, connection, name)

	if logging.HasError(ctx) {
		if login.Id != "" {
			logging.AddError(
				ctx,
				"Login already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_login.<name> %s", login.Id))
		}
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
	connection := r.ConnectionCache.Connect(ctx, connectionId, true, false)

	if connection.ConnectionResourceStatus == sql.ConnectionResourceStatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if logging.HasError(ctx) {
		return
	}

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

	if logging.HasError(ctx) {
		return
	}

	if login.Id == "" {
		resp.State.RemoveResource(ctx)
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

	connection := r.ConnectionCache.Connect(ctx, state.Server.ValueString(), true, false)

	if connection.ConnectionResourceStatus == sql.ConnectionResourceStatusNotFound {
		return
	}

	sql.DropLogin(ctx, connection, state.Sid.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping user failed", fmt.Sprintf("Dropping user %s failed", state.Name.ValueString()))
	}
}

func (r *SQLLoginResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	login := sql.ParseLoginId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, login.Connection, true, true)

	if logging.HasError(ctx) {
		return
	}

	login = sql.GetLoginFromSid(ctx, connection, login.Sid)

	if logging.HasError(ctx) {
		return
	}

	state := SQLLoginResourceModel{
		Id:     types.StringValue(login.Id),
		Server: types.StringValue(login.Connection),
		Name:   types.StringValue(login.Name),
		Sid:    types.StringValue(login.Sid),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
