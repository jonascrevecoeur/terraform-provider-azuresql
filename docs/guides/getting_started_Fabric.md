# Getting started - Fabric

With `azuresql` you can manage the data plane (schemas, roles, ...) for Fabric warehouses and the SQL endpoint of Fabric lakehouses.

This guide gets you started with using `azuresql` to manage your database. It covers
* Provider initialization
* Provisioning users
* Managing schemas
* Granting permissions

## Provider initialization
azuresql authenticates to Fabric via the  [Azure default credential chain](https://microsoft.github.io/spring-cloud-azure/4.0.0-beta.3/4.0.0-beta.3/reference/html/authentication.html) using the active credentials of the terminal running Terraform. 
When you are using the provider locally, ensure you are logged into the Azure CLI, such that the provider can use these credentials
```
az login
```
when running non-interactively the necessary credentials can be configured via a managed identity or service principal. 

-> The identity using the provider requires the admin permission on the warehouse/lakehouse managed via azuresql

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
Since azuresql connects to an Azure subscription, we must still reference via data sources each database used by the provider. To connect, you need to lookup the SQL connection string of your warehouse in Fabric. This connection string takes the form
`<identifier>.datawarehouse.fabric.microsoft.com`. The \<identifier> being a random combination of letters, numbers and special characters. 

The snippet below references the database named `mywarehouse` on the server with connections string `<identifier>.datawarehouse.fabric.microsoft.com`

```
data "azuresql_fabricworkspace" "server" {
  server = "<identifier>"
}

data "azuresql_database" "database" {
  server    = data.azuresql_fabricworkspace.server.id
  name      = "mywarehouse"
}
```

Every other azuresql resource or data source takes a parameter `database` to determine in which Fabric warehouse or lakehouse the resource should be managed.

## Provisioning users
Fabric allows you to register users directly via their EntraID name. 

```
resource "azuresql_user" "myUser" {
  database       = data.azuresql_database.database.id
  name           = "myUser@mycompany.com"
  authentication = "AzureAD"
}
```
!> Only users assigned to the Fabric workspace can be added to the database using this resource

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


