package fabricworkspace

import "github.com/hashicorp/terraform-plugin-framework/types"

type fabricworkspaceDataSourceModel struct {
	ConnectionId types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
}
