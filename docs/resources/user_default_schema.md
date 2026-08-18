# azuresql_user_default_schema (Resource)

Manages the `DEFAULT_SCHEMA` property of a database user.

## Why this resource exists

SQL Server stores a user's default schema on the database principal. It is
not a schema-level permission and it cannot be expressed by an
`azuresql_permission` or `azuresql_role_assignment` resource. Before this
resource, callers had to use the `azuresql_execute_sql` data source to run
`ALTER USER ... WITH DEFAULT_SCHEMA ...`, which made the setting an
untyped, apply-time side effect and made drift difficult to detect.

This resource makes the setting part of Terraform state. It reads the current
value during refresh, updates it when `default_schema` changes, and resets it
to `dbo` when the resource is destroyed. The referenced schema must already
exist; setting a default schema does not grant any SQL permissions.

## Example Usage

```terraform
resource "azuresql_user" "ds_ops" {
  database       = data.azuresql_database.target.id
  name           = "my-ds-ops-group"
  authentication = "AzureAD"
}

resource "azuresql_user_default_schema" "ds_ops" {
  database       = data.azuresql_database.target.id
  user           = azuresql_user.ds_ops.id
  default_schema = "dbo"
}
```

## Argument Reference

The following arguments are supported:

* `database` - (Required) ID of the database containing the user.
* `user` - (Required) ID of an `azuresql_user` in the same database.
* `default_schema` - (Required) Name of the existing schema to use as the
  user's default schema.

## Import

Import using the database connection ID and the database principal ID:

```shell
terraform import azuresql_user_default_schema.ds_ops \
  'sqlserver::my-server:1433:my-database/user_default_schema/42'
```
