package provider

import (
	"context"
	"terraform-provider-azuresql/internal/services/database"
	"terraform-provider-azuresql/internal/services/database_scoped_credential"
	"terraform-provider-azuresql/internal/services/execute_sql"
	"terraform-provider-azuresql/internal/services/external_data_source"
	"terraform-provider-azuresql/internal/services/function"
	"terraform-provider-azuresql/internal/services/master_key"
	"terraform-provider-azuresql/internal/services/permission"
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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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
			" azuresql authenticates using the [Azure default credential chain](https://learn.microsoft.com/en-us/dotnet/api/azure.identity.defaultazurecredential)." +
			" By authentiation to Azure instead of a specific database/server instance, the provider can be used to manage multiple SQL databases/servers at once." +
			"\\\n\\\n" +
			"The identitiy using this providers requires full control on the database/server to be configured."}
}

func (p *azuresql_provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	cache := sql.NewCache()

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
	}
}
