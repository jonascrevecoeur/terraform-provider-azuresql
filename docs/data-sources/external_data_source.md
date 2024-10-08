---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "azuresql_external_data_source Data Source - terraform-provider-azuresql"
subcategory: ""
description: |-
  Read  external data sources.
---

# azuresql_external_data_source (Data Source)

Read  external data sources.

**Supported**: `SQL Database`, `Synapse serverless database` 

**Not supported**: `SQL Server`, `Synapse dedicated database`, `Fabric`

!> This resource is under development. At the moment only blobstorage is supported as an external data source.

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

data "azuresql_external_data_source" "example" {
  database    = data.azuresql_database.database.id
  name        = "mysource"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Argument reference
The following arguments are supported:

- `database` (Required, String) The ID of the database in which the external data source is registered.

- `name` (Required, String) Name of the external data source.


### Attributes Reference
In addition to the arguments listed above, the following read only attributes are exported:

- `location` (String) Location (path) of the external data source.
- `credential` (String) Id of the credential (`azuresql_database_scoped_credential`) used to access the external data source. 
- `id` (String) The azuresql ID of the external data source resource.
- `data_source_id` (Number) ID of the external data source in the database.

## ID structure

The ID is formed as `<database>`/externaldatasource/`<data source id>`, where
* `<database>`  is the ID of the `azuresql_database` resource.
* `<data_source_id>` is the id of the data source in the database. It can be found by running `select data_source_id from sys.external_data_sources where name = '<data source name>'`.
