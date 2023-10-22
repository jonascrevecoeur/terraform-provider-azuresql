---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "azuresql_role Data Source - terraform-provider-azuresql"
subcategory: ""
description: |-
  Read sql server and database roles.
---

# azuresql_role (Data Source)

Read sql server and database roles.

**Supported**: `SQL Server`, `SQL Database`, `Synapse serverless server`, `Synapse serverless database` 

**Not supported**: `Synapse dedicated server`, `Synapse dedicated database`


## Example Usage

```terraform
provider "azuresql" {
}

data "azuresql_synapseserver" "server" {
  server = "myserver"
}

data "azuresql_database" "database" {
  server = data.azuresql_synapseserver.server.id
  name   = "mydatabase"
}

data "azuresql_role" "myrole" {
    database       = azuresql_database.database.id
    name           = "myrole"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `database` (String) The ID of the database in which the role exists. 
- `server` (String) Id of the server where the role exists.

-> Only one of `database` or `server` should be specified.

- `name` (String) Name of the SQL role.

### Read-Only

- `id` (String) The azuresql ID of the role resource.
- `owner` (String) ID of the role or user owning this role.
- `principal_id` (Number) Principal ID of the role in the database.