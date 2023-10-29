package permission

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PermissionDataSourceModel struct {
	Id          types.String   `tfsdk:"id"`
	Database    types.String   `tfsdk:"database"`
	Server      types.String   `tfsdk:"server"`
	Scope       types.String   `tfsdk:"scope"`
	Principal   types.String   `tfsdk:"principal"`
	Permissions []types.String `tfsdk:"permissions"`
}

type PermissionResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Database   types.String `tfsdk:"database"`
	Server     types.String `tfsdk:"server"`
	Scope      types.String `tfsdk:"scope"`
	Principal  types.String `tfsdk:"principal"`
	Permission types.String `tfsdk:"permission"`
	Action     types.String `tfsdk:"action"`
}
