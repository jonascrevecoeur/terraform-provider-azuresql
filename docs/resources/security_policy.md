---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "azuresql_security_policy Resource - terraform-provider-azuresql"
subcategory: ""
description: |-
  Manage a database security policy.
---

# azuresql_security_policy (Resource)

Manage a database security policy.

**Supported**: `SQL database`, `Fabric`

**Not supported**: `Synapse dedicated database`, `Synapse serverless database`


## Example Usage

```terraform
provider "azuresql" {
}

data "azuresql_sqlserver" "server" {
  server = "mysqlserver"
}

data "azuresql_database" "database" {
  server = data.azuresql_sqlserver.server.id
  name   = "mydatabase"
}

data "azuresql_schema" "dbo" {
    database  = data.azuresql_database.database.id
    name      = "dbo"
}

resource "azuresql_function" "filter" {
    database  = data.azuresql_database.database.id
    name      = "filter"
    schema    = data.azuresql_schema.dbo.id
    raw       = <<-EOT
        create function dbo.filter(@user as varchar(50))
        returns table 
        with SCHEMABINDING AS
        return  select 1 as result
        where @user in (select suser_sname())
    EOT
}

resource "azuresql_security_policy" "filter" {
  database = data.azuresql_database.database.id
  name     = "filter_policy"
  schema   = data.azuresql_schema.dbo.id
}

data "azuresql_table" "mytable" {
    database    = data.azuresql_database.database.id
    schema      =  data.azuresql_schema.dbo.id
    name        = "mytable"
}

resource "azuresql_security_predicate" "filter_select" {
  database          = data.azuresql_database.database.id
  security_policy   = azuresql_security_policy.filter.id
  table             = data.azuresql_table.mytable.id
  rule              = "dbo.filter(user)"
  type              = "filter"

  # a lifecyle rule informs Terraform about the dependency between this predicate and the function
  lifecycle {
    replace_triggered_by = [azuresql_function.filter]
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Argument reference
The following arguments are supported:

- `database` (String) ID of the database where the security policy should be created.
- `name` (String) Name of the security policy
- `schema` (String) ID of the `azuresql_schema` in which the security policy should be created.

### Attributes Reference
In addition to the arguments listed above, the following read only attributes are exported:


- `id` (String)  azuresql ID of the security policy resource.
- `object_id` (Number) ID of the security policy object in the database.

## ID structure

The ID is formed as `<database>`/securitypolicy/`<object id>`, where
* `<database>` is the ID of the `azuresql_database` resource.
* `<object id>` is the id of the security policy in the database. It can be found by running `select object_id('<schema name>.<security policy  name>')`.

## Import

You can import a security policy using 

```shell
terraform import azuresql_security_policy.<resource name> <id>
```