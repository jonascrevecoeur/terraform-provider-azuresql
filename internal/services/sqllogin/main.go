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

type SQLLoginPasswordPropertiesResourceModel struct {
	Length             types.Int32  `tfsdk:"length"`
	AllowedSpecialChar types.String `tfsdk:"allowed_special_characters"`
	MinSpecialChar     types.Int32  `tfsdk:"min_special_characters"`
	MinNum             types.Int32  `tfsdk:"min_numbers"`
	MinUpperCaseum     types.Int32  `tfsdk:"min_uppercase"`
}

type SQLLoginResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Server   types.String `tfsdk:"server"`
	Name     types.String `tfsdk:"name"`
	Password types.Object `tfsdk:"password"`
	Sid      types.String `tfsdk:"sid"`
}
