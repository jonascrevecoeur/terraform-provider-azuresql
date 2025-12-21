resource "azurerm_resource_group" "this" {
  name     = "rg-azuresql"
  location = local.region
}

data "azuread_client_config" "me" {}
