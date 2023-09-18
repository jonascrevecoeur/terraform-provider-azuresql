package sqlserver

import "github.com/hashicorp/terraform-plugin-framework/types"

type sqlserverDataSourceModel struct {
	ConnectionId types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Port         types.Int64  `tfsdk:"port"`
}
