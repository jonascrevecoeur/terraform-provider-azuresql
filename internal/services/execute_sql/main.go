package execute_sql

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ExecuteSQLDataSourceModel struct {
	Database types.String `tfsdk:"database"`
	Server   types.String `tfsdk:"server"`
	SQL      types.String `tfsdk:"sql"`
}
