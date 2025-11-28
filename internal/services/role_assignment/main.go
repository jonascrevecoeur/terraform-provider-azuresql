package role_assignment

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type RoleAssignmentResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Database  types.String `tfsdk:"database"`
	Server    types.String `tfsdk:"server"`
	Role      types.String `tfsdk:"role"`
	Principal types.String `tfsdk:"principal"`
}
