terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "=4.1.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "3.7.0"
    }
  }
}

provider "azurerm" {
  subscription_id = local.azure_subscription
  features {}
}

provider "azuread" {
  tenant_id = local.tenant_id
  use_cli   = true
}
