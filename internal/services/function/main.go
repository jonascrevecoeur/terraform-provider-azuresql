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

type FunctionPropertiesResourceModel struct {
	Arguments     []FunctionArgumentResourceModel `tfsdk:"arguments"`
	ReturnType    types.String                    `tfsdk:"return_type"`
	Executor      types.String                    `tfsdk:"executor"`
	Schemabinding types.Bool                      `tfsdk:"schemabinding"`
	Definition    types.String                    `tfsdk:"definition"`
}
type FunctionArgumentResourceModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type FunctionResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Database   types.String `tfsdk:"database"`
	ObjectId   types.Int64  `tfsdk:"object_id"`
	Name       types.String `tfsdk:"name"`
	Schema     types.String `tfsdk:"schema"`
	Properites types.Object `tfsdk:"properties"`
	Raw        types.String `tfsdk:"raw"`
}
