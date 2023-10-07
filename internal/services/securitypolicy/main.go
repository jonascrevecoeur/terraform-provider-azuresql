package securitypolicy

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type SecurityPolicyDataSourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	Name     types.String `tfsdk:"name"`
	ObjectId types.Int64  `tfsdk:"object_id"`
	Schema   types.String `tfsdk:"schema"`
}

type SecurityPolicyResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	Name     types.String `tfsdk:"name"`
	ObjectId types.Int64  `tfsdk:"object_id"`
	Schema   types.String `tfsdk:"schema"`
}
