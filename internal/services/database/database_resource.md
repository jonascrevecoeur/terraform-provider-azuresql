---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "azuresql_database Resource - terraform-provider-azuresql"
subcategory: ""
description: |-
  Manage databases
---

# azuresql_database (Resource)

Manage the lifecycle of a Synapse database.

**Supported**: `Synapse serverless` 

**Not supported**: `SQL Server`, `Synapse dedicated`, `Fabric`

~> To avoid accidental deletion of the database it is highly recommended that you use the `prevent_destroy` lifecycle argument in configuring this resource. For more information see the [terraform documentation](https://developer.hashicorp.com/terraform/tutorials/state/resource-lifecycle#prevent-resource-deletion)

~> This resource does not support managing databases in sql server. USe the `azurerm_mssql_database` resource in the `azurerm` provider instead. For more information see the [azurerm_mssql_database documentation](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/resources/mssql_database)

## Example Usage

```terraform
data "azuresql_synapseserver" "server" {
  server = "mysynapseserver"
}

resource "azuresql_database" "database" {
  server  = data.azuresql_synapseserver.server
  name    = "example"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Argument reference
The following arguments are supported:

- `name` (Required, String) Name of the database within the server.
- `server` (Required, String) Id of the `azuresql_synapseserver` resource.

### Attributes Reference
In addition to the arguments listed above, the following read only attributes are exported:

- `id` (String) ID of the sqlserver connection in `azuresql`. This ID is passed to other `azuresql` resources and data sources to indicate that the resource should be created in/read from this database, respectively. 

## ID structure

The ID is formed as `<server>:<name>`, where
* `<server>` is the ID of the `azuresql_synapseserver` resource.
* `<name>` is the name of the database.

## Import

You can import a database using 

```shell
terraform import azuresql_dabase.<resource name> <id>
```