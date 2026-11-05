package provider

import (
	"context"
	"terraform-provider-azuresql/internal/auth"
	"terraform-provider-azuresql/internal/services/database"
	"terraform-provider-azuresql/internal/services/database_scoped_credential"
	"terraform-provider-azuresql/internal/services/execute_sql"
	"terraform-provider-azuresql/internal/services/external_data_source"
	"terraform-provider-azuresql/internal/services/fabricworkspace"
	"terraform-provider-azuresql/internal/services/function"
	"terraform-provider-azuresql/internal/services/master_key"
	"terraform-provider-azuresql/internal/services/permission"
	"terraform-provider-azuresql/internal/services/procedure"
	"terraform-provider-azuresql/internal/services/role"
	"terraform-provider-azuresql/internal/services/role_assignment"
	dbschema "terraform-provider-azuresql/internal/services/schema"
	"terraform-provider-azuresql/internal/services/securitypolicy"
	"terraform-provider-azuresql/internal/services/securitypredicate"
	login "terraform-provider-azuresql/internal/services/sqllogin"
	"terraform-provider-azuresql/internal/services/sqlserver"
	"terraform-provider-azuresql/internal/services/synapseserver"
	"terraform-provider-azuresql/internal/services/table"
	"terraform-provider-azuresql/internal/services/user"
	"terraform-provider-azuresql/internal/services/view"
	"terraform-provider-azuresql/internal/sql"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &azuresql_provider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &azuresql_provider{
			version: version,
		}
	}
}

type azuresql_provider struct {
	version string
}

// Metadata returns the provider type name.
func (p *azuresql_provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "azuresql"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *azuresql_provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The azuresql provider can be used to configure SQL resources in `Azure SQL server`, `Azure SQL database` and in `AzureSynapse serverless pool`." +
			" azuresql authenticates using the [Azure default credential chain](https://learn.microsoft.com/en-us/dotnet/api/azure.identity.defaultazurecredential)," +
			" unless a client secret, client certificate, OIDC or managed identity authentication is configured." +
			" By authentiation to Azure instead of a specific database/server instance, the provider can be used to manage multiple SQL databases/servers at once." +
			"\\\n\\\n" +
			"The identitiy using this providers requires full control on the database/server to be configured.",
		Attributes: map[string]schema.Attribute{
			"subscription_id": schema.StringAttribute{
				Optional: true,
			},
			"check_server_exists": schema.BoolAttribute{
				Optional: true,
				Validators: []validator.Bool{
					boolvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("subscription_id"),
					}...),
				},
			},
			"check_database_exists": schema.BoolAttribute{
				Optional: true,
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "The client (application) ID to authenticate as. Environment variable: `ARM_CLIENT_ID`, only read when an authentication method is configured.",
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Description: "The Entra ID tenant to authenticate against. Environment variable: `ARM_TENANT_ID`, only read when an authentication method is configured.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Authenticate as a service principal using a client secret. Environment variable: `ARM_CLIENT_SECRET`.",
			},
			"client_certificate": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Authenticate as a service principal using a base64 encoded PKCS#12 client certificate. Environment variable: `ARM_CLIENT_CERTIFICATE`.",
			},
			"client_certificate_path": schema.StringAttribute{
				Optional:    true,
				Description: "Authenticate as a service principal using the PKCS#12 client certificate at this path. Environment variable: `ARM_CLIENT_CERTIFICATE_PATH`.",
			},
			"client_certificate_password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The password of the client certificate. Environment variable: `ARM_CLIENT_CERTIFICATE_PASSWORD`.",
			},
			"use_oidc": schema.BoolAttribute{
				Optional: true,
				Description: "Authenticate using OIDC (workload identity federation): an OIDC token issued by an external identity provider is exchanged for an Entra ID token." +
					" The OIDC token is read from the environment: supplied directly (`ARM_OIDC_TOKEN`) or as a file (`ARM_OIDC_TOKEN_FILE_PATH`), for example a Kubernetes service account token," +
					" or requested from a token endpoint (`ARM_OIDC_REQUEST_URL`, `ACTIONS_ID_TOKEN_REQUEST_URL` or `SYSTEM_OIDCREQUESTURI`," +
					" with `ARM_OIDC_REQUEST_TOKEN`, `ACTIONS_ID_TOKEN_REQUEST_TOKEN` or `SYSTEM_ACCESSTOKEN`) in Azure DevOps pipelines" +
					" (with `ARM_ADO_PIPELINE_SERVICE_CONNECTION_ID` or `ARM_OIDC_AZURE_SERVICE_CONNECTION_ID`) and GitHub Actions. Environment variable: `ARM_USE_OIDC`.",
			},
			"use_msi": schema.BoolAttribute{
				Optional:    true,
				Description: "Authenticate using a managed identity. Set `client_id` to select a user-assigned identity. Environment variable: `ARM_USE_MSI`.",
			},
		},
	}
}

func (p *azuresql_provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ProviderConfigModel
	resp.Diagnostics.Append(
		req.Config.Get(ctx, &config)...,
	)

	var check_database_exists bool
	if config.CheckDatabaseExists.IsNull() {
		check_database_exists = true
	} else {
		check_database_exists = config.CheckDatabaseExists.ValueBool()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	authCfg, err := authConfig(config)
	if err != nil {
		resp.Diagnostics.AddError("Invalid authentication configuration", err.Error())
		return
	}
	credential, err := auth.NewCredential(ctx, authCfg)
	if err != nil {
		resp.Diagnostics.AddError("Invalid authentication configuration", err.Error())
		return
	}

	cache := sql.NewCache(
		config.SubscriptionId.ValueString(),
		config.CheckServerExists.ValueBool(),
		check_database_exists)
	cache.Credential = credential

	resp.DataSourceData = &cache
	resp.ResourceData = &cache
}

// DataSources defines the data sources implemented in the provider.
func (p *azuresql_provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		sqlserver.NewServerDataSource,
		synapseserver.NewSynapseServerDataSource,
		login.NewSQLLoginDataSource,
		user.NewUserDataSource,
		role.NewRoleDataSource,
		dbschema.NewSchemaDataSource,
		database.NewDatabaseDataSource,
		permission.NewPermissionDataSource,
		function.NewFunctionDataSource,
		table.NewTableDataSource,
		execute_sql.NewExecuteSQLDataSource,
		external_data_source.NewExternalDataSourceDataSource,
		view.NewViewDataSource,
		procedure.NewProcedureDataSource,
		fabricworkspace.NewFabricWorkspaceDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *azuresql_provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		user.NewUserResource,
		login.NewSQLLoginResource,
		role.NewRoleResource,
		dbschema.NewSchemaResource,
		permission.NewPermissionResource,
		function.NewFunctionResource,
		securitypolicy.NewSecurityPolicyResource,
		securitypredicate.NewSecurityPredicateResource,
		role_assignment.NewRoleAssignmentResource,
		master_key.NewMasterKeyResource,
		database_scoped_credential.NewDatabaseScopedCredentialResource,
		external_data_source.NewExternalDataSourceResource,
		view.NewViewResource,
		database.NewDatabaseResource,
		procedure.NewProcedureResource,
	}
}
