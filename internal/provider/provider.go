package provider

import (
	"context"
	"terraform-provider-azuresql/internal/services/database"
	"terraform-provider-azuresql/internal/services/function"
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
		Description: "azuresql enables you to simultaneously connect to multiple Azure sql servers, Azure sql databases and Azure Synapse resources. " +
			"No arguments are passed when setting up the provider, instead the provider uses AzureDefaultCredential passthrough to connect to any Azure sql resource specified."}
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
	}
}
