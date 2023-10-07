package function

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
	_ resource.Resource                = &FunctionResource{}
	_ resource.ResourceWithConfigure   = &FunctionResource{}
	_ resource.ResourceWithImportState = &FunctionResource{}
)

func NewFunctionResource() resource.Resource {
	return &FunctionResource{}
}

type FunctionResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *FunctionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_function"
}

func (r *FunctionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database or server user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "Id of the database where the user should be created. database or server should be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the function",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Id of the function object in the database",
			},
			"schema": schema.StringAttribute{
				Required:    true,
				Description: "Schema where the function resides.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"raw": schema.StringAttribute{
				Required:    true,
				Description: "Raw definition of the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *FunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan FunctionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	database := plan.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	function := sql.CreateFunctionFromDefinition(ctx, connection, name, plan.Schema.ValueString(), plan.Raw.ValueString())

	if logging.HasError(ctx) {
		return
	}

	plan.Id = types.StringValue(function.Id)
	plan.ObjectId = types.Int64Value(function.ObjectId)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *FunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state FunctionResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	function := sql.GetFunctionFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) || function.Id == "" {
		return
	}

	state.Name = types.StringValue(function.Name)
	state.ObjectId = types.Int64Value(function.ObjectId)
	state.Raw = types.StringValue(function.Raw)
	state.Schema = types.StringValue(function.Schema)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *FunctionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *FunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *FunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state FunctionResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	sql.DropFunction(ctx, connection, state.Id.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping function failed", fmt.Sprintf("Dropping function %s failed", state.Name.ValueString()))
	}
}

func (r *FunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	/*ctx = utils.WithDiagnostics(ctx, &resp.Diagnostics)

	user := sql._user_parse_id(ctx, req.ID)

	if utils.HasError(ctx) {
		return
	}

	resp.State.SetAttribute(ctx, path.Root("connection_string"), user.ConnectionString)
	resp.State.SetAttribute(ctx, path.Root("principal_id"), user.PrincipalId)*/
}
