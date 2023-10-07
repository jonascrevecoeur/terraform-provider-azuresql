package function

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type FunctionDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	ObjectId types.Int64  `tfsdk:"object_id"`
	Name     types.String `tfsdk:"name"`
	Schema   types.String `tfsdk:"schema"`
	Raw      types.String `tfsdk:"raw"`
}

type FunctionResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	ObjectId types.Int64  `tfsdk:"object_id"`
	Name     types.String `tfsdk:"name"`
	Schema   types.String `tfsdk:"schema"`
	Raw      types.String `tfsdk:"raw"`
}
