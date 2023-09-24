package role

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleDataSourceModel struct {
	Id          types.String `tfsdk:"id"`
	Database    types.String `tfsdk:"database"`
	Server      types.String `tfsdk:"server"`
	Name        types.String `tfsdk:"name"`
	PrincipalId types.Int64  `tfsdk:"principal_id"`
	Owner       types.String `tfsdk:"owner"`
}

type RoleResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Database    types.String `tfsdk:"database"`
	Server      types.String `tfsdk:"server"`
	Name        types.String `tfsdk:"name"`
	PrincipalId types.Int64  `tfsdk:"principal_id"`
	Owner       types.String `tfsdk:"owner"`
}
