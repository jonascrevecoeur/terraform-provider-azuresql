# Getting started - Synapse serverless

With `azuresql` you can manage the data plane (schemas, roles, ...) for Azure Synapse. Where applicable resources can be created both at the server and the database level.

~> Only serverless pools are supported by the provider

This guide gets you started with using `azuresql` to manage your database. It covers
* Provider initialization
* Provisioning users
* Managing schemas
* Granting permissions

## Provider initialization
azuresql authenticates to Synapse via the  [Azure default credential chain](https://microsoft.github.io/spring-cloud-azure/4.0.0-beta.3/4.0.0-beta.3/reference/html/authentication.html) using the active credentials of the terminal running Terraform. 
When you are using the provider locally, ensure you are logged into the Azure CLI, such that the provider can use these credentials
```
az login
```
when running non-interactively the necessary credentials can be configured via a managed identity or service principal. 

-> The identity using the provider requires the synapse administrator role on the workspace managed via azuresql

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
Since azuresql connects to an Azure subscription, we must still reference via data sources each database used by the provider. The snippet below references the database named `mydatabase` on `mysynapseworkspace`. 

```
data "azuresql_synapseserver" "server" {
  server = "mysynapseserver"
}

data "azuresql_database" "database" {
  server    = data.azuresql_synapseserver.server.id
  name      = "mydatabase"
}
```

Every other azuresql resource or data source takes a parameter `database` or `server` to determine in which Synapse server or database the resource should be managed.

## Provisioning users
Synapse offers different types of users
* Users with an associated SQL login
* External users authenticating through Azure AD
These two types can be provisioned via the `azuresql_user` resource.

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
You can register users directly via their EntraID name. The identity using the provider requires the `read.all` permission on EntraID to import the users.

```
resource "azuresql_user" "myUser" {
  database       = data.azuresql_database.database.id
  name           = "myUser@mycompany.com"
  authentication = "AzureAD"
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


