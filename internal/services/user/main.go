package user

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type UserDataSourceModel struct {
	Id             types.String `tfsdk:"id"`
	Database       types.String `tfsdk:"database"`
	Server         types.String `tfsdk:"server"`
	Name           types.String `tfsdk:"name"`
	PrincipalId    types.Int64  `tfsdk:"principal_id"`
	Authentication types.String `tfsdk:"authentication"`
	Type           types.String `tfsdk:"type"`
	Sid            types.String `tfsdk:"sid"`
}

type UserResourceModel struct {
	Id             types.String `tfsdk:"id"`
	Database       types.String `tfsdk:"database"`
	Server         types.String `tfsdk:"server"`
	Name           types.String `tfsdk:"name"`
	PrincipalId    types.Int64  `tfsdk:"principal_id"`
	Authentication types.String `tfsdk:"authentication"`
	Type           types.String `tfsdk:"type"`
	Login          types.String `tfsdk:"login"`
	ObjectID       types.String `tfsdk:"object_id"`
	Sid            types.String `tfsdk:"sid"`
}
