package procedure

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
	_ resource.Resource                = &ProcedureResource{}
	_ resource.ResourceWithConfigure   = &ProcedureResource{}
	_ resource.ResourceWithImportState = &ProcedureResource{}
)

func NewProcedureResource() resource.Resource {
	return &ProcedureResource{}
}

type ProcedureResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *ProcedureResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_procedure"
}

func (r *ProcedureResource) SchemaPropertiesAttributes() map[string]attr.Type {
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

func (r *ProcedureResource) SchemaProperties() map[string]schema.Attribute {
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
		"executor": schema.StringAttribute{
			Optional: true,
			Computed: true,
			Default:  stringdefault.StaticString("caller"),
		},
		"definition": schema.StringAttribute{
			Required: true,
		},
	}
}

func (r *ProcedureResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Description: "Name of the procedure",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"object_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Id of the procedure object in the database",
			},
			"schema": schema.StringAttribute{
				Required:    true,
				Description: "Schema where the procedure resides.",
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
				Description: "Raw definition of the procedure.",
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

func GetProcedureProps(rm *ProcedurePropertiesResourceModel) (props sql.ProcedureProps) {
	if rm == nil {
		return sql.ProcedureProps{}
	} else {
		var arguments []sql.ProcedureArgument
		for _, argument := range rm.Arguments {
			arguments = append(arguments, sql.ProcedureArgument{
				Name: argument.Name.ValueString(),
				Type: argument.Type.ValueString(),
			})
		}
		return sql.ProcedureProps{
			Arguments:     arguments,
			Definition:    rm.Definition.ValueString(),
			Executor:      rm.Executor.ValueString(),
			Schemabinding: false,
		}
	}
}

func ProcedurePropsToResourceModel(props sql.ProcedureProps) ProcedurePropertiesResourceModel {

	rm := ProcedurePropertiesResourceModel{
		Executor:   types.StringValue(props.Executor),
		Definition: types.StringValue(props.Definition),
	}

	rm.Arguments = []ProcedureArgumentResourceModel{}
	for _, argument := range props.Arguments {
		rm.Arguments = append(rm.Arguments, ProcedureArgumentResourceModel{
			Name: types.StringValue(argument.Name),
			Type: types.StringValue(argument.Type),
		})
	}

	return rm
}

func (r *ProcedureResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan ProcedureResourceModel

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

	var procedure sql.Procedure
	if plan.Properites.IsNull() || plan.Properites.IsUnknown() {
		procedure = sql.CreateProcedureFromRaw(ctx, connection, name, plan.Schema.ValueString(), plan.Raw.ValueString())
	} else {
		var planProps ProcedurePropertiesResourceModel
		diags := req.Plan.GetAttribute(ctx, path.Root("properties"), &planProps)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		procedure = sql.CreateProcedureFromProperties(ctx, connection, name, plan.Schema.ValueString(), GetProcedureProps(&planProps))
	}

	if logging.HasError(ctx) {
		if procedure.Id != "" {
			logging.AddError(
				ctx,
				"Procedure already exists",
				fmt.Sprintf("You can import this resource using `terraform import azuresql_procedure.<name> %s", procedure.Id))
		}
		return
	}

	plan.Id = types.StringValue(procedure.Id)
	plan.ObjectId = types.Int64Value(procedure.ObjectId)
	plan.Raw = types.StringValue(procedure.Raw)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: reenable when setting procedure properties again to computed
	/*
		diags = resp.State.SetAttribute(ctx, path.Root("properties"), ProcedurePropsToResourceModel(procedure.Properties))
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}*/
}

func (r *ProcedureResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state ProcedureResourceModel
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

	procedure := sql.GetProcedureFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) {
		return
	}

	if procedure.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(procedure.Name)
	state.ObjectId = types.Int64Value(procedure.ObjectId)
	state.Raw = types.StringValue(procedure.Raw)
	state.Schema = types.StringValue(procedure.Schema)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *ProcedureResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *ProcedureResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	logging.AddError(ctx, "Update not implemented", "Update should never be called for procedure resources as any change requires a delete and recreate.")
}

func (r *ProcedureResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state ProcedureResourceModel

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

	sql.DropProcedure(ctx, connection, state.Id.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping procedure failed", fmt.Sprintf("Dropping procedure %s failed", state.Name.ValueString()))
	}
}

func (r *ProcedureResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)
	tflog.Info(ctx, fmt.Sprintf("Importing procedure %s", req.ID))

	procedure := sql.ParseProcedureId(ctx, req.ID)

	if logging.HasError(ctx) {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, procedure.Connection, false, true)

	if logging.HasError(ctx) {
		return
	}

	procedure = sql.GetProcedureFromId(ctx, connection, req.ID, true)

	if logging.HasError(ctx) {
		return
	}

	state := ProcedureResourceModel{
		Id:         types.StringValue(procedure.Id),
		Database:   types.StringValue(procedure.Connection),
		ObjectId:   types.Int64Value(procedure.ObjectId),
		Name:       types.StringValue(procedure.Name),
		Schema:     types.StringValue(procedure.Schema),
		Raw:        types.StringValue(procedure.Raw),
		Properites: types.ObjectNull(r.SchemaPropertiesAttributes()),
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
