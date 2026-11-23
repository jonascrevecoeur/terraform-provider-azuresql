package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProviderConfigModel struct {
	SubscriptionId      types.String `tfsdk:"subscription_id"`
	CheckServerExists   types.Bool   `tfsdk:"check_server_exists"`
	CheckDatabaseExists types.Bool   `tfsdk:"check_database_exists"`

	ClientId                  types.String `tfsdk:"client_id"`
	TenantId                  types.String `tfsdk:"tenant_id"`
	ClientSecret              types.String `tfsdk:"client_secret"`
	ClientCertificate         types.String `tfsdk:"client_certificate"`
	ClientCertificatePath     types.String `tfsdk:"client_certificate_path"`
	ClientCertificatePassword types.String `tfsdk:"client_certificate_password"`
	UseOIDC                   types.Bool   `tfsdk:"use_oidc"`
	UseMSI                    types.Bool   `tfsdk:"use_msi"`
}
