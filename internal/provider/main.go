package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ProviderConfigModel struct {
	SubscriptionId      types.String `tfsdk:"subscription_id"`
	CheckServerExists   types.Bool   `tfsdk:"check_server_exists"`
	CheckDatabaseExists types.Bool   `tfsdk:"check_database_exists"`
}
