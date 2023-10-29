package database_scoped_credential

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DatabaseScopedCredentialResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Database     types.String `tfsdk:"database"`
	Name         types.String `tfsdk:"name"`
	Identity     types.String `tfsdk:"identity"`
	Secret       types.String `tfsdk:"secret"`
	CredentialId types.Int64  `tfsdk:"credential_id"`
}
