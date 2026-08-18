package user_default_schema

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

var (
	_ resource.Resource                = &UserDefaultSchemaResource{}
	_ resource.ResourceWithConfigure   = &UserDefaultSchemaResource{}
	_ resource.ResourceWithImportState = &UserDefaultSchemaResource{}
)

func NewUserDefaultSchemaResource() resource.Resource {
	return &UserDefaultSchemaResource{}
}

type UserDefaultSchemaResource struct {
	ConnectionCache *sql.ConnectionCache
}

func (r *UserDefaultSchemaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_default_schema"
}

func (r *UserDefaultSchemaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Default schema assigned to a SQL database user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for Terraform used to import the resource.",
			},
			"database": schema.StringAttribute{
				Required:    true,
				Description: "ID of the database containing the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user": schema.StringAttribute{
				Required:    true,
				Description: "Azure SQL provider ID of the user whose default schema is managed.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"default_schema": schema.StringAttribute{
				Required:    true,
				Description: "Name of the schema used as the user's default schema. The schema must already exist.",
			},
		},
	}
}

func (r *UserDefaultSchemaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var plan UserDefaultSchemaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, plan.Database.ValueString(), false, true)
	if logging.HasError(ctx) {
		return
	}

	setting := sql.GetUserDefaultSchema(ctx, connection, plan.User.ValueString(), true)
	if logging.HasError(ctx) {
		return
	}

	sql.SetUserDefaultSchema(ctx, connection, setting.UserName, plan.DefaultSchema.ValueString())
	if logging.HasError(ctx) {
		return
	}

	setting = sql.GetUserDefaultSchema(ctx, connection, setting.User, true)
	if logging.HasError(ctx) {
		return
	}
	plan.Id = types.StringValue(setting.Id)
	plan.User = types.StringValue(setting.User)
	plan.DefaultSchema = types.StringValue(setting.DefaultSchema)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *UserDefaultSchemaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state UserDefaultSchemaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, state.Database.ValueString(), false, false)
	if logging.HasError(ctx) {
		return
	}
	if connection.ConnectionResourceStatus == sql.ConnectionResourceStatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	setting := sql.GetUserDefaultSchema(ctx, connection, state.User.ValueString(), false)
	if logging.HasError(ctx) {
		return
	}
	if setting.Id == "" {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Id = types.StringValue(setting.Id)
	state.User = types.StringValue(setting.User)
	state.DefaultSchema = types.StringValue(setting.DefaultSchema)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UserDefaultSchemaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state UserDefaultSchemaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var plan UserDefaultSchemaResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, state.Database.ValueString(), false, true)
	if logging.HasError(ctx) {
		return
	}

	setting := sql.GetUserDefaultSchema(ctx, connection, state.User.ValueString(), true)
	if logging.HasError(ctx) {
		return
	}
	sql.SetUserDefaultSchema(ctx, connection, setting.UserName, plan.DefaultSchema.ValueString())
	if logging.HasError(ctx) {
		return
	}

	state.DefaultSchema = plan.DefaultSchema
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UserDefaultSchemaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	var state UserDefaultSchemaResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, state.Database.ValueString(), false, false)
	if logging.HasError(ctx) {
		return
	}
	if connection.ConnectionResourceStatus == sql.ConnectionResourceStatusNotFound {
		return
	}

	setting := sql.GetUserDefaultSchema(ctx, connection, state.User.ValueString(), false)
	if logging.HasError(ctx) || setting.Id == "" {
		return
	}
	sql.SetUserDefaultSchema(ctx, connection, setting.UserName, "dbo")
	if logging.HasError(ctx) {
		resp.Diagnostics.AddError("Resetting default schema failed", fmt.Sprintf("Resetting default schema for user %s failed", setting.UserName))
	}
}

func (r *UserDefaultSchemaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cache, ok := req.ProviderData.(*sql.ConnectionCache)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *sql.ConnectionCache, got: %T.", req.ProviderData),
		)
		return
	}
	r.ConnectionCache = cache
}

func (r *UserDefaultSchemaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	ctx = logging.WithDiagnostics(ctx, &resp.Diagnostics)

	setting := sql.ParseUserDefaultSchemaId(ctx, req.ID)
	if logging.HasError(ctx) {
		return
	}

	connection := r.ConnectionCache.Connect(ctx, setting.Connection, false, true)
	if logging.HasError(ctx) {
		return
	}
	setting = sql.GetUserDefaultSchema(ctx, connection, setting.User, true)
	if logging.HasError(ctx) {
		return
	}

	state := UserDefaultSchemaResourceModel{
		Id:            types.StringValue(setting.Id),
		Database:      types.StringValue(setting.Connection),
		User:          types.StringValue(setting.User),
		DefaultSchema: types.StringValue(setting.DefaultSchema),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
