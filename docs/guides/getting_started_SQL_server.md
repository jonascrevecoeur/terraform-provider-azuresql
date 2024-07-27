# Getting started - Azure SQL server

With `azuresql` you can manage the data plane (schemas, roles, ...) for Azure SQL server. Where applicable resources can be created both at the server and the database level.

This guide gets you started with using `azuresql` to manage your database. It covers
* Provider initialization
* Provisioning users
* Managing schemas
* Granting permissions
* Escape hatch

## Provider initialization
azuresql authenticates to SQL server via the  [Azure default credential chain](https://microsoft.github.io/spring-cloud-azure/4.0.0-beta.3/4.0.0-beta.3/reference/html/authentication.html) using the active credentials of the terminal running Terraform. 
When you are using the provider locally, ensure you are logged into the Azure CLI, such that the provider can use these credentials
```
az login
```
when running non-interactively the necessary credentials can be configured via a managed identity or service principal. 

-> The identity using the provider requires admin access to the databases managed via the provider.

Since credentials are taken from the active terminal, you don't define authentication details in the provider. azuresql can be initialized with an empty provider block. 
```
terraform {
  required_providers {
    azuresql = {
      source  = "jonascrevecoeur/azuresql"
    }
  }
}

provider "azuresql" {}
```

## Establishing a database connection
Since azuresql connects to an Azure subscription, we must still reference via data sources each database used by the provider. The snippet below references the database named `mydatabase` on `mysqlserver`. 

```
data "azuresql_sqlserver" "server" {
  server = "mysqlserver"
}

data "azuresql_database" "database" {
  server    = data.azuresql_sqlserver.server.id
  name      = "mydatabase"
}
```

Every other azuresql resource or data source takes a parameter `database` or `server` to determine in which SQL server or database the resource should be managed.

## Provisioning users
SQL server offers different types of users
* Users with an associated SQL login
* External users authenticating through Azure AD
* Users without login
These three types can be provisioned via the `azuresql_user` resource.

### User with a SQL login
When you use azuresql to create a login, a secure password is automatically generated which can be retrieved via the `password` attribute of the `azuresql_login` resource. Note that the login resource exists at the server level, whereas the user is created at the database level.
```
resource "azuresql_login" "login" {
    server   = data.azuresql_sqlserver.server.id
    name     = "mylogin"
}

resource "azuresql_user" "user" {
    database        = data.azuresql_database.database.id
    name            = "myuser"
    authentication  = "SQLLogin"
    login           = azuresql_login.login.id
}
```

### Azure AD / EntraId authentication
When [EntraID authentication](https://learn.microsoft.com/en-us/azure/azure-sql/database/authentication-aad-configure?view=azuresql&tabs=azure-powershell#provision-azure-ad-admin-sql-database) is enabled for your SQL server, you can register users directly via their EntraID name. 

```
resource "azuresql_user" "myUser" {
  database       = data.azuresql_database.database.id
  name           = "myUser@mycompany.com"
  authentication = "AzureAD"
}
```

### Users without login
You can also create database accounts without a login. These account can only be used via `execute as`
```
resource "azuresql_user" "myUser" {
  database       = data.azuresql_database.database.id
  name           = "myuser"
  authentication = "WithoutLogin"
}
```

## Managing schemas

The snippet below creates a new schema named example
```
resource "azuresql_schema" "schema" {
    database     = data.azuresql_database.database.id
    name         = "example"
}
```
~> azuresql will not destroy a schema containing resources (e.g. tables) not managed by the provider. You need to manually delete these resources if the provider has to remove the schema.

## Granting permissions
The `azuresql_permission` resource is used to grant or deny permissions to database principals. To define a permission you need to specify
* Principal: The user or role receiving the permission
* Scope: The object (database, schema, table, ...) on which the permission is given
* permission: The permission to grant

The block below creates a user John Smith with CREATE TABLE Permission at the database level and the select permission on a schema named example.
```
resource "azuresql_user" "JohnSmith" {
  database       = data.azuresql_database.database.id
  name           = "johnsmith@mycompany.com"
  authentication = "AzureAD"
}

resource "azuresql_permission" "johnsmith_select_example" {
    database       = data.azuresql_database.database.id
    scope          = data.azuresql_database.database.id
	  principal      = data.azuresql_user.JohnSmith.id
    permission     = "create table"
}

data "azuresql_schema" "example" {
	database       = data.azuresql_database.database.id
  name           = "example"
}

resource "azuresql_permission" "johnsmith_select_example" {
    database       = data.azuresql_database.database.id
    scope          = data.azuresql_schema.example.id
	  principal      = data.azuresql_user.JohnSmith.id
    permission     = "select"
}
```

## Escape hatch 
The provider does not offer complete coverage of the T-SQL language used in Azure SQL Server. For instance, tables can be referenced via data sources, but cannot be created via the provider. If you want to ensure the table exists, you can use the `azuresql_execute_sql` data source as an escape hatch
```
data "azuresql_execute_sql" "test" {
  database  = data.azuresql_database.database.id
  sql       = <<-EOT
    IF OBJECT_ID('mytable', 'U') IS NULL
    create table mytable(
      col1 float
    )
  EOT
}
```

This SQL statement will be executed each time the provider runs. It will create the table, but won't manage its lifecycle. 

If the escape hatch does not meet your needs, you can submit a feature request on our [GitHub repository](https://github.com/jonascrevecoeur/terraform-provider-azuresql).


