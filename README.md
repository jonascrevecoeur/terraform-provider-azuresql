# Terraform Provider for Azure SQL resources

This provider lets you manage the data plane of Azure SQL resources. At the moment it supports:
* Azure SQL server
* Azure SQL database
* Azure Synapse serverless pool

The provider enables passwordless authentiation through the [Azure default credential chain](https://learn.microsoft.com/en-us/dotnet/api/azure.identity.defaultazurecredential). This enables you to manage multiple SQL resources using a single provider block.

* Documentation: https://registry.terraform.io/providers/jonascrevecoeur/azuresql/latest/docs

## Usage

```terraform
provider "azuresql" {
}


# Use data statements to load the server and database to be managed using the provider
data "azuresql_sqlserver" "server" {
  server = "mysqlserver"
  port   = 1433
}

data "azuresql_database" "database" {
  server = data.azuresql_sqlserver.server.id
  name   = "mydatabase"
}

# create a user in the database
resource "azuresql_user" "test" {
  # every resource/datasource uses the database/server argument to determine where to create the resource
  database       = azuresql_database.database.id
  name           = "myuser"
  authentication = "WithoutLogin"
}
```