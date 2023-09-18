package login

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SQLLoginDataSourceModel struct {
	Id     types.String `tfsdk:"id"`
	Server types.String `tfsdk:"server"`
	Name   types.String `tfsdk:"name"`
	Sid    types.String `tfsdk:"sid"`
}

type SQLLoginResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Server   types.String `tfsdk:"server"`
	Name     types.String `tfsdk:"name"`
	Password types.String `tfsdk:"password"`
	Sid      types.String `tfsdk:"sid"`
}
