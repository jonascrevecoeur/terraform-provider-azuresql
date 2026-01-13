package function

import (
	"context"
	"fmt"

	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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

func (r *FunctionResource) SchemaPropertiesAttributes() map[string]attr.Type {
	return map[string]attr.Type{
		"arguments": types.ListType{
			ElemType: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name": types.StringType,
					"type": types.StringType,
				},
			},
		},
		"return_type":   types.StringType,
		"definition":    types.StringType,
		"schemabinding": types.BoolType,
		"executor":      types.StringType,
	}
}

func (r *FunctionResource) SchemaProperties() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"arguments": schema.ListNestedAttribute{
			Optional: true,
			NestedObject: schema.NestedAttributeObject{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Required: true,
					},
					"type": schema.StringAttribute{
						Required: true,
					},
				},
			},
		},
		"return_type": schema.StringAttribute{
			Description: "Type of the returned value.",
			Required:    true,
		},
		"executor": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("caller"),
		},
		"schemabinding": schema.BoolAttribute{
			Optional: true,
			Computed: true,
			Default:  booldefault.StaticBool(true),
			Description: `If set to true, prevents the referenced objects to be changed in any way that would break 
			the functionality of the function.`,
		},
		"definition": schema.StringAttribute{
			Required: true,
		},
	}
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
			"properties": schema.SingleNestedAttribute{
				Optional: true,
				// Disabled for now - requires proper parsing of raw to props
				// Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: r.SchemaProperties(),
			},
			"raw": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Raw definition of the function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.Expressions{
						path.MatchRoot("properties"),
					}...),
				},
			},
		},
	}
}

func GetFunctionProps(rm *FunctionPropertiesResourceModel) (props sql.FunctionProps) {
	if rm == nil {
		return sql.FunctionProps{}
	} else {
		var arguments []sql.FunctionArgument
		for _, argument := range rm.Arguments {
			arguments = append(arguments, sql.FunctionArgument{
				Name: argument.Name.ValueString(),
				Type: argument.Type.ValueString(),
			})
		}
		return sql.FunctionProps{
			Arguments:     arguments,
			ReturnType:    rm.ReturnType.ValueString(),
			Definition:    rm.Definition.ValueString(),
			Executor:      rm.Executor.ValueString(),
			Schemabinding: rm.Schemabinding.ValueBool(),
		}
	}
}

func FunctionPropsToResourceModel(props sql.FunctionProps) FunctionPropertiesResourceModel {

	rm := FunctionPropertiesResourceModel{
		ReturnType:    types.StringValue(props.ReturnType),
		Executor:      types.StringValue(props.Executor),
		Schemabinding: types.BoolValue(props.Schemabinding),
		Definition:    types.StringValue(props.Definition),
	}

	rm.Arguments = []FunctionArgumentResourceModel{}
	for _, argument := range props.Arguments {
		rm.Arguments = append(rm.Arguments, FunctionArgumentResourceModel{
			Name: types.StringValue(argument.Name),
			Type: types.StringValue(argument.Type),
		})
	}

	return rm
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
	connection := r.ConnectionCache.Connect(ctx, database, false, true)

	if logging.HasError(ctx) {
		return
	}

	if connection.Provider == "synapsededicated" {
		logging.AddError(ctx, "Invalid config",
			"`azuresql_function` resource is not support on Synapse dedicated.")
		return
	}

	var function sql.Function
	var raw string
	if plan.Properites.IsNull() || plan.Properites.IsUnknown() {
		function = sql.CreateFunctionFromRaw(ctx, connection, name, plan.Schema.ValueString(), plan.Raw.ValueString())
		raw = plan.Raw.ValueString()
	} else {
		var planProps FunctionPropertiesResourceModel
		diags := req.Plan.GetAttribute(ctx, path.Root("properties"), &planProps)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		function = sql.CreateFunctionFromProperties(ctx, connection, name, plan.Schema.ValueString(), GetFunctionProps(&planProps))
		raw = function.Raw
	}

	if logging.HasError(ctx) {
		if function.Id != "" {
			logging.AddError(
				ctx,
				"Function already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_function.<name> %s", function.Id))
		}
		return
	}

	plan.Id = types.StringValue(function.Id)
	plan.ObjectId = types.Int64Value(function.ObjectId)
	plan.Raw = types.StringValue(raw)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: reenable when setting function properties again to computed
	/*
		diags = resp.State.SetAttribute(ctx, path.Root("properties"), FunctionPropsToResourceModel(function.Properties))
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}*/
}

func (r *FunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state FunctionResourceModel
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

	function := sql.GetFunctionFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) {
		return
	}

	if function.Id == "" {
		resp.State.RemoveResource(ctx)
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
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	logging.AddError(ctx, "Update not implemented", "Update should never be called for function resources as any change requires a delete and recreate.")
}

func (r *FunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state FunctionResourceModel

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

	sql.DropFunction(ctx, connection, state.Id.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping function failed", fmt.Sprintf("Dropping function %s failed", state.Name.ValueString()))
	}
}

func (r *FunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	tflog.Info(ctx, fmt.Sprintf("Importing function %s", req.ID))

	function := sql.ParseFunctionId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, function.Connection, false, true)

	if logging.HasError(ctx) {
		return
	}

	function = sql.GetFunctionFromId(ctx, connection, req.ID, true)

	if logging.HasError(ctx) {
		return
	}

	state := FunctionResourceModel{
		Id:         types.StringValue(function.Id),
		Database:   types.StringValue(function.Connection),
		ObjectId:   types.Int64Value(function.ObjectId),
		Name:       types.StringValue(function.Name),
		Schema:     types.StringValue(function.Schema),
		Raw:        types.StringValue(function.Raw),
		Properites: types.ObjectNull(r.SchemaPropertiesAttributes()),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
