---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "azuresql_synapseserver Data Source - terraform-provider-azuresql"
subcategory: ""
description: |-
  Define a synapseserver connection to be used by the `azuresql` provider.
---

# azuresql_synapseserver (Data Source)

Defines a connection to the synapse server. Creating the data source does not yet open/test the connection. Opening the connection happens when it is used for reading/provisioning other `azuresql` resources.

**Supported**: `Synapse serverless server`

**Not supported**: `Synapse dedicated server`

## Example Usage

```terraform
data "azuresql_synapseserver" "server" {
  server = "mysynapseserver"
  port   = 1433
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Argument reference
The following arguments are supported:

- `name` (Required, String) Name of the Synapse server. This is the value in the url preceeding `-ondemand.sql.azuresynapse.net`
  
- `port` (Optional, Number) Port through which to connect to the synapse server (default 1433).

### Attributes Reference
In addition to the arguments listed above, the following read only attributes are exported:

- `id` (String) ID of the synapse connection in `azuresql`. This ID is passed to other `azuresql` resources and data sources to indicate that the resource should be created in/read from this server, respectively.

## ID structure

The ID is formed as `synapseserver::<name>:<port>`, where
* `<name>` is the name of the server.
* `<port>` is the port of the server. Default 1433.