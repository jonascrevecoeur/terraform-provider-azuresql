package database_scoped_credential_test

import (
	"fmt"
	"testing"

	"terraform-provider-azuresql/internal/acceptance"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type DatabaseScopedCredentialResource struct{}

func TestAccCreateDatabaseScopedCredential(t *testing.T) {
	r := DatabaseScopedCredentialResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.DatabaseConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.basic(connection, data.RandomString, "SHARED ACCESS SIGNATURE", "secret"+data.RandomString),
				},
			},
		})
	})
}

func TestAccImportDatabaseScopedCredential(t *testing.T) {
	// Testing import with a database scoped credential without secret, as the secret is always blank after import
	r := DatabaseScopedCredentialResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.DatabaseConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.without_secret(connection, data.RandomString, "SHARED ACCESS SIGNATURE"),
				},
				{
					Config:            r.without_secret(connection, data.RandomString, "SHARED ACCESS SIGNATURE"),
					ResourceName:      "azuresql_database_scoped_credential.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccUpdateCredential(t *testing.T) {
	r := DatabaseScopedCredentialResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.DatabaseConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.basic(connection, data.RandomString, "SHARED ACCESS SIGNATURE", "secret"+data.RandomString),
					Check: resource.TestCheckResourceAttr("azuresql_database_scoped_credential.test", "secret", "secret"+data.RandomString),
				},
				{
					Config: r.basic(connection, data.RandomString, "SHARED ACCESS SIGNATURE", "secret2"+data.RandomString),
					Check: resource.TestCheckResourceAttr("azuresql_database_scoped_credential.test", "secret", "secret2"+data.RandomString),
				},
			},
		})
	})
}

func (r DatabaseScopedCredentialResource) basic(connection string, name string, identity string, secret string) string {
	return fmt.Sprintf(`
	%[1]s

	resource "azuresql_database_scoped_credential" "test" {
		database 	= "%[2]s"
		name  		= "tfcredential_%[3]s"
		identity  	= "%[4]s"
		secret  	= "%[5]s"

		depends_on = [azuresql_master_key.test]
	}
`, r.template(connection), connection, name, identity, secret)
}

func (r DatabaseScopedCredentialResource) without_secret(connection string, name string, identity string) string {
	return fmt.Sprintf(`
	%[1]s

	resource "azuresql_database_scoped_credential" "test" {
		database 	= "%[2]s"
		name  		= "tfcredential_%[3]s"
		identity  	= "%[4]s"

		depends_on = [azuresql_master_key.test]
	}
`, r.template(connection), connection, name, identity)
}

func (r DatabaseScopedCredentialResource) template(connection string) string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}

		resource "azuresql_master_key" "test" {
			database 	= "%[1]s"
		}
	`, connection)
}
