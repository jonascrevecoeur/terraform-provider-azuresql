package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SchemaDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	Name     types.String `tfsdk:"name"`
	Owner    types.String `tfsdk:"owner"`
	SchemaId types.Int64  `tfsdk:"schema_id"`
}

type SchemaResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	Name     types.String `tfsdk:"name"`
	Owner    types.String `tfsdk:"owner"`
	SchemaId types.Int64  `tfsdk:"schema_id"`
}
