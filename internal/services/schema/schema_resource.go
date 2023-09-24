package schema

import (
	"context"
	"fmt"

	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &SchemaResource{}
	_ resource.ResourceWithConfigure   = &SchemaResource{}
	_ resource.ResourceWithImportState = &SchemaResource{}
)

func NewSchemaResource() resource.Resource {
	return &SchemaResource{}
}

type SchemaResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *SchemaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schema"
}

func (r *SchemaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database schema.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the database where the schema exists.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the schema.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Schema ID of the schema in the database.",
			},
			"owner": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Principal owning the schema.",
			},
		},
	}
}

func (r SchemaResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	// no modification required on create or delete
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var state SchemaResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	// in Synapse serverless alter authorization cannot be used
	// -> a replace is required when owner changes
	if connection.Provider == "synapse" {
		resp.RequiresReplace.Append(path.Root("owner"))
	}
}

func (r *SchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan SchemaResourceModel
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

	owner := plan.Owner.ValueString()

	schema := sql.CreateSchema(ctx, connection, name, owner)

	if logging.HasError(ctx) {
		return
	}

	plan.Id = types.StringValue(schema.Id)
	plan.SchemaId = types.Int64Value(schema.SchemaId)
	plan.Owner = types.StringValue(schema.Owner)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state SchemaResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	schema := sql.GetSchemaFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) || schema.Id == "" {
		return
	}

	state.SchemaId = types.Int64Value(schema.SchemaId)
	state.Name = types.StringValue(schema.Name)
	state.Owner = types.StringValue(schema.Owner)
	state.Id = types.StringValue(schema.Id)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SchemaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *SchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state SchemaResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan SchemaResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// changes in these values would have triggered a replacement
	// so they are identical for state/plan
	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)
	id := state.Id.ValueString()

	// update owner
	if !state.Owner.Equal(plan.Owner) {
		sql.UpdateSchemaOwner(ctx, connection, id, plan.Owner.ValueString())

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

func (r *SchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state SchemaResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	sql.DropSchema(ctx, connection, state.SchemaId.ValueInt64())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping schema failed", fmt.Sprintf("Dropping schema %s failed", state.Name.ValueString()))
	}
}

func (r *SchemaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	/*ctx = utils.WithDiagnostics(ctx, &resp.Diagnostics)

	user := sql._user_parse_id(ctx, req.ID)

	if utils.HasError(ctx) {
		return
	}

	resp.State.SetAttribute(ctx, path.Root("connection_string"), user.ConnectionString)
	resp.State.SetAttribute(ctx, path.Root("principal_id"), user.PrincipalId)*/
}
