data "azuresql_sqlserver" "server" {
  server = "mysqlserver"
  port   = 1433
}
