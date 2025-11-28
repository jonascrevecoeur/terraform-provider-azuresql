# Configuring Authentication in CI/CD

The `azuresql` provider uses the [Azure Default Credential Chain](https://learn.microsoft.com/en-us/dotnet/api/azure.identity.defaultazurecredential) to obtain an access token for the server or database.

In non-interactive environments such as CI/CD pipelines, it is typically necessary to manually provide credentials via environment variables.

## Authentication Using Client ID and Client Secret

To authenticate using a service principal (client ID and client secret), set the following environment variables:

- `AZURE_CLIENT_ID`
- `AZURE_TENANT_ID`
- `AZURE_CLIENT_SECRET`

## Authentication Using Workload Identity Federation

To authenticate using workload identity federation (OIDC), set the following environment variables:

- `AZURE_CLIENT_ID`
- `AZURE_TENANT_ID`
- `ID_TOKEN`
- `AZURE_FEDERATED_TOKEN_FILE`

---

# Using a Service Connection in Azure DevOps

In Azure DevOps, you can retrieve the necessary credentials from a Service Connection and make them available to your Terraform tasks.

To set the required environment variables, add the following task to your pipeline **before** running any Terraform commands:

```yaml
- task: AzureCLI@2
  name: set_variables
  displayName: "Set Terraform Credentials"
  inputs:
    azureSubscription: "${{ parameters.service_connection_ARM }}"
    addSpnToEnvironment: true
    scriptLocation: inlineScript
    scriptType: "bash"
    inlineScript: |
      echo $idToken > /.idtoken
      echo "##vso[task.setvariable variable=AZURE_CLIENT_ID]$servicePrincipalId"
      echo "##vso[task.setvariable variable=AZURE_TENANT_ID]$tenantId"
      echo "##vso[task.setvariable variable=AZURE_CLIENT_SECRET]$servicePrincipalKey"
      echo "##vso[task.setvariable variable=ID_TOKEN]$idToken"
      echo "##vso[task.setvariable variable=AZURE_FEDERATED_TOKEN_FILE]/.idtoken"
```

This task will set the necessary credentials independent of whether the connection was created via a client-secret or federated credential.