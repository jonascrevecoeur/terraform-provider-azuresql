package procedure

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProcedureDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	ObjectId types.Int64  `tfsdk:"object_id"`
	Name     types.String `tfsdk:"name"`
	Schema   types.String `tfsdk:"schema"`
	Raw      types.String `tfsdk:"raw"`
}

type ProcedurePropertiesResourceModel struct {
	Arguments  []ProcedureArgumentResourceModel `tfsdk:"arguments"`
	Executor   types.String                     `tfsdk:"executor"`
	Definition types.String                     `tfsdk:"definition"`
}
type ProcedureArgumentResourceModel struct {
	Name types.String `tfsdk:"name"`
	Type types.String `tfsdk:"type"`
}

type ProcedureResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Database   types.String `tfsdk:"database"`
	ObjectId   types.Int64  `tfsdk:"object_id"`
	Name       types.String `tfsdk:"name"`
	Schema     types.String `tfsdk:"schema"`
	Properites types.Object `tfsdk:"properties"`
	Raw        types.String `tfsdk:"raw"`
}
