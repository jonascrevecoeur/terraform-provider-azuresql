package view

import (
	"context"
	"fmt"

	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ViewResource{}
	_ resource.ResourceWithConfigure   = &ViewResource{}
	_ resource.ResourceWithImportState = &ViewResource{}
)

func NewViewResource() resource.Resource {
	return &ViewResource{}
}

type ViewResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *ViewResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_view"
}

func (r *ViewResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Database view.",
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
				Description: "Name of the view",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Id of the view object in the database",
			},
			"schema": schema.StringAttribute{
				Required:    true,
				Description: "Schema where the view resides.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schemabinding": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
				Description: `If set to true, prevents the referenced objects to be changed in any way that would break 
				the functionality of the function.`,
			},
			"check_option": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: `If true, all data modification statements in the view have to match the select statements. [(official docs)](https://learn.microsoft.com/en-us/sql/t-sql/statements/create-view-transact-sql?view=sql-server-ver16#check-option)`,
			},
			"definition": schema.StringAttribute{
				Required:    true,
				Description: "Definition of the view.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *ViewResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan ViewResourceModel
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

	view := sql.CreateViewFromDefinition(ctx, connection, name, plan.Schema.ValueString(), plan.Definition.ValueString(), plan.Schemabinding.ValueBool(), plan.CheckOption.ValueBool())

	if logging.HasError(ctx) {
		if view.Id != "" {
			logging.AddError(
				ctx,
				"View already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_view.<name> %s", view.Id))
		}
		return
	}

	plan.Id = types.StringValue(view.Id)
	plan.ObjectId = types.Int64Value(view.ObjectId)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ViewResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state ViewResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	view := sql.GetViewFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) {
		return
	}

	if view.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(view.Name)
	state.ObjectId = types.Int64Value(view.ObjectId)

	if !sql.IsViewDefinitionEquivalent(ctx, state.Definition.ValueString(), view.Definition) {
		state.Definition = types.StringValue(view.Definition)
	}

	state.Schemabinding = types.BoolValue(view.Schemabinding)
	state.CheckOption = types.BoolValue(view.CheckOption)

	state.Schema = types.StringValue(view.Schema)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ViewResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *ViewResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *ViewResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state ViewResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	sql.DropView(ctx, connection, state.Id.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping view failed", fmt.Sprintf("Dropping view %s failed", state.Name.ValueString()))
	}
}

func (r *ViewResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	tflog.Info(ctx, fmt.Sprintf("Importing view %s", req.ID))

	view := sql.ParseViewId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, view.Connection, false)

	if logging.HasError(ctx) {
		return
	}

	view = sql.GetViewFromId(ctx, connection, req.ID, true)

	if logging.HasError(ctx) {
		return
	}

	state := ViewResourceModel{
		Id:         types.StringValue(view.Id),
		Database:   types.StringValue(view.Connection),
		ObjectId:   types.Int64Value(view.ObjectId),
		Name:       types.StringValue(view.Name),
		Schema:     types.StringValue(view.Schema),
		Definition: types.StringValue(view.Definition),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
