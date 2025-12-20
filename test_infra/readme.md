Creating the infrastructure requires a locals.tf file with the following content:

```{hcl}
locals {
  random_string      = ""
  azure_subscription = ""
  tenant_id          = ""
  region             = ""
  user               = ""
  group              = ""
}
```

Manual actions required:

- Create a database in the Synapse server
- Create a fabric capacity + lakehouse