---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "azuresql_fabricworkspace Data Source - terraform-provider-azuresql"
subcategory: ""
description: |-
  Define a fabric workspace connection to be used by the `azuresql` provider.
---

# azuresql_fabricworkspace (Data Source)

Defines a connection to a fabrick workspace. Creating the data source does not yet open/test the connection. Opening the connection happens when it is used for reading/provisioning other `azuresql` resources.

~> When connecting to Fabric workspaces, you have to set `check_server_exists` to `false` when creating the `azuresql` provider. At the moment, Fabric doesn't offer an API to efficiencly determine whether a SQL endpoint exists. 

## Example Usage

```terraform
data "azuresql_fabricworkspace" "workspace" {
  workspace = "myfabricworkspace"
  port   = 1433
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Argument reference
The following arguments are supported:

- `endpoint` (Required, String) SQL endpoint of the Fabric workspace. This is the value in the connection string preceeding `.datawarehouse.fabric.microsoft.com`
  
### Attributes Reference
In addition to the arguments listed above, the following read only attributes are exported:

- `id` (String) ID of the workspace connection in `azuresql`. This ID is passed to other `azuresql` resources and data sources to indicate that the resource should be created in/read from this workspace, respectively.

## ID structure

The ID is formed as `fabric::<name>`, where
* `<name>` is the name of the workspace.