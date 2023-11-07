---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "azuresql_view Resource - terraform-provider-azuresql"
subcategory: ""
description: |-
  Manage database views.
---

# azuresql_view (Resource)

Manage database views.

**Supported**: `SQL Database`, `Synapse serverless database` 

**Not supported**: `Synapse dedicated database`


## Example Usage

```terraform
provider "azuresql" {
}

data "azuresql_sqlserver" "server" {
  server  = "mysqlserver"
}

data "azuresql_database" "database" {
  server  = data.azuresql_sqlserver.server.id
  name    = "mydatabase"
}

data "azuresql_schema" "dbo" {
    database  = data.azuresql_database.database.id
    name      = "dbo"
}

resource "azuresql_view" "example" {
    database    = data.azuresql_database.database.id
    name        = "example"
    schema      = data.azuresql_schema.dbo.id
    definition  = <<-EOT
        select * from mytable
    EOT
}

```

~> Hint: Since the view is created using a raw query, Terraform might not automatically detect all dependencies on other azuresql resources (e.g. other schemas, and tables mentioned in the view). You can resolve this by manually specifying these dependencies in a `replace_triggered_by` lifecycle rule for the `azuresql_view` resource.

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `database` (String) ID of the database where the view should be created.
- `name` (String) Name of the view
- `schema` (String) ID of the `azuresql_schema` in which the view should be created.
- `definition` (String) SQL statement executed by the view


### Read-Only

- `id` (String)  azuresql ID of the view resource.
- `object_id` (Number) ID of the view object in the database

## ID structure

The ID is formed as `<database>`/view/`<object_id>`, where
* `<database>` is the ID of the `azuresql_database` resource.
* `<object_id>` is the id of the view in the database. It can be found by running `select object_id('<schema>.<view name>')`.

## Import

You can import a view using 

```terraform import azuresql_view.<resource name> <id>```