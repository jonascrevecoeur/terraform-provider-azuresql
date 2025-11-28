package master_key

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type MasterKeyResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Database types.String `tfsdk:"database"`
	Password types.String `tfsdk:"password"`
}
