resource "random_password" "sqlserver" {
  length           = 16
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
}

resource "azurerm_mssql_server" "this" {
  name                         = "azuresql-${local.random_string}"
  resource_group_name          = azurerm_resource_group.this.name
  location                     = azurerm_resource_group.this.location
  version                      = "12.0"
  administrator_login          = "azuresqladmin"
  administrator_login_password = random_password.sqlserver.result
  minimum_tls_version          = "1.2"

  azuread_administrator {
    login_username = local.user
    object_id      = data.azuread_client_config.me.object_id
  }
}

resource "azurerm_mssql_firewall_rule" "allow_all" {
  name             = "allowall"
  server_id        = azurerm_mssql_server.this.id
  start_ip_address = "0.0.0.0"
  end_ip_address   = "255.255.255.255"
}

resource "azurerm_mssql_database" "core" {
  name                        = "core"
  server_id                   = azurerm_mssql_server.this.id
  collation                   = "SQL_Latin1_General_CP1_CI_AS"
  max_size_gb                 = 1
  sku_name                    = "GP_S_Gen5_1"
  storage_account_type        = "Local"
  min_capacity                = 0.5
  auto_pause_delay_in_minutes = -1
}
