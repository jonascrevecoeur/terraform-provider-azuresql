package view

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ViewResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Database   types.String `tfsdk:"database"`
	ObjectId   types.Int64  `tfsdk:"object_id"`
	Name       types.String `tfsdk:"name"`
	Schema     types.String `tfsdk:"schema"`
	Definition types.String `tfsdk:"definition"`
}
