package table

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type TableDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	Name     types.String `tfsdk:"name"`
	Schema   types.String `tfsdk:"schema"`
	ObjectId types.Int64  `tfsdk:"object_id"`
}
