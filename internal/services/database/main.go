package database

import "github.com/hashicorp/terraform-plugin-framework/types"

type databaseDataSourceModel struct {
	ConnectionId types.String `tfsdk:"id"`
	Server       types.String `tfsdk:"server"`
	Name         types.String `tfsdk:"name"`
}
