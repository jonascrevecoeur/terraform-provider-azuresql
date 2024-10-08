---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "azuresql_database Data Source - terraform-provider-azuresql"
subcategory: ""
description: |-
  Define a database connection to be used by the `azuresql` provider.
---

# azuresql_database (Data Source)

Defines a connection to a database. Creating the data source does not yet open/test the connection. Opening the connection happens when reading/provisioning other `azuresql` resources.

**Supported**: `SQL Database`, `Synapse serverless database`, `Fabric`

**Not supported**: `Synapse dedicated database`


## Example Usage

```terraform
data "azuresql_sqlserver" "sqlserver" {
  server = "mysqlserver"
}

data "azuresql_synapseserver" "synapseserver" {
  server = "mysynapseserver"
}

# connect to a database on the sqlserver
data "azuresql_database" "sqldatabase" {
  server = data.azuresql_sqlserver.sqlserver.id
  name   = "mydatabase"
}

# connect to a database on the synapse server
data "azuresql_database" "synapsedatabase" {
  server = data.azuresql_synapseserver.synapseserver.id
  name   = "mydatabase"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Argument reference
The following arguments are supported:

- `name` (Required, String) Name of the database within the server.
- `server` (Required, String) Id of the `azuresql_sqlserver` or `azuresql_synapseserver` resource.

### Attributes Reference
In addition to the arguments listed above, the following read only attributes are exported:

- `id` (String) ID of the sqlserver connection in `azuresql`. This ID is passed to other `azuresql` resources and data sources to indicate that the resource should be created in/read from this database, respectively. 

## ID structure

The ID is formed as `<server>:<name>`, where
* `<server>` is the ID of the `azuresql_sqlserver` or `azuresql_synapseserver` resource.
* `<name>` is the name of the database.