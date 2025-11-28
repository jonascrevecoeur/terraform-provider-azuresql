provider "azuresql" {
}


# Use data statements to load the server and database to be managed using the provider
data "azuresql_sqlserver" "server" {
  name = "mysqlserver"
  port = 1433
}

data "azuresql_database" "database" {
  server = data.azuresql_sqlserver.server.id
  name   = "mydatabase"
}

# create a user in the database
resource "azuresql_user" "test" {
  # every resource/datasource uses the database/server argument to determine where to create the resource
  database       = data.azuresql_database.database.id
  name           = "myuser"
  authentication = "WithoutLogin"
}
