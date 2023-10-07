package securitypolicy

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
	_ resource.Resource                = &SecurityPolicyResource{}
	_ resource.ResourceWithConfigure   = &SecurityPolicyResource{}
	_ resource.ResourceWithImportState = &SecurityPolicyResource{}
)

func NewSecurityPolicyResource() resource.Resource {
	return &SecurityPolicyResource{}
}

type SecurityPolicyResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *SecurityPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_policy"
}

func (r *SecurityPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SQL database or server user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Optional:    true,
				Description: "Id of the database where the user should be created. database or server should be specified.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the user",
			},
			"object_id": schema.Int64Attribute{
				Computed:    true,
				Description: "Principal ID of the user in the database.",
			},
			"schema": schema.StringAttribute{
				Required:    true,
				Description: "SecurityPolicy or user owning the securityPolicy.",
			},
		},
	}
}

func (r *SecurityPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan SecurityPolicyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	schema := plan.Schema.ValueString()
	database := plan.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	if connection.Provider == "synapse" {
		logging.AddError(ctx, "Invalid config", "Security policies are not supported on Synapse.")
		return
	}

	securityPolicy := sql.CreateSecurityPolicy(ctx, connection, name, schema)

	if logging.HasError(ctx) {
		return
	}

	plan.Id = types.StringValue(securityPolicy.Id)
	plan.ObjectId = types.Int64Value(securityPolicy.ObjectId)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SecurityPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state SecurityPolicyResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	securityPolicy := sql.GetSecurityPolicyFromId(ctx, connection, state.Id.ValueString(), false)

	if logging.HasError(ctx) || securityPolicy.Id == "" {
		return
	}

	state.ObjectId = types.Int64Value(securityPolicy.ObjectId)
	state.Name = types.StringValue(securityPolicy.Name)
	state.Schema = types.StringValue(securityPolicy.Schema)
	state.Id = types.StringValue(securityPolicy.Id)

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SecurityPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {

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

func (r *SecurityPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

func (r *SecurityPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state SecurityPolicyResourceModel

	// Read input configured in data block
	resp.Diagnostics.Append(
		req.State.Get(ctx, &state)...,
	)

	database := state.Database.ValueString()
	connection := r.ConnectionCache.Connect(ctx, database, false)

	if logging.HasError(ctx) {
		return
	}

	sql.DropSecurityPolicy(ctx, connection, state.Id.ValueString())

	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Dropping securityPolicy failed", fmt.Sprintf("Dropping securityPolicy %s failed", state.Name.ValueString()))
	}
}

func (r *SecurityPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {

	/*ctx = utils.WithDiagnostics(ctx, &resp.Diagnostics)

	user := sql._user_parse_id(ctx, req.ID)

	if utils.HasError(ctx) {
		return
	}

	resp.State.SetAttribute(ctx, path.Root("connection_string"), user.ConnectionString)
	resp.State.SetAttribute(ctx, path.Root("principal_id"), user.PrincipalId)*/
}
