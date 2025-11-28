package schema_test

import (
	"fmt"
	"regexp"
	"testing"

	"terraform-provider-azuresql/internal/acceptance"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SchemaResource struct{}

func TestAccCreateSchemaBasic(t *testing.T) {
	r := SchemaResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
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
					Config: r.basic(connection, data.RandomString),
				},
				{
					Config:            r.basic(connection, data.RandomString),
					ResourceName:      "azuresql_schema.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccCreateDuplicateResource(t *testing.T) {
	r := SchemaResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
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
					Config:      r.duplicate_schema(connection, data.RandomString),
					ExpectError: regexp.MustCompile("You can import this resource using"),
				},
			},
		})
	})
}

func TestAccCreateSchemaWithOwner(t *testing.T) {
	r := SchemaResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
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
					Config: r.withOwner(connection, data.RandomString),
				},
			},
		})
	})
}

func TestAccUpdateSchemaOwner(t *testing.T) {
	r := SchemaResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
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
					Config: r.updateOwner(connection, data.RandomString, 1),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("azuresql_schema.test", "owner", "azuresql_role.owner1", "id"),
					),
				},
				{
					Config: r.updateOwner(connection, data.RandomString, 2),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair("azuresql_schema.test", "owner", "azuresql_role.owner2", "id"),
					),
				},
			},
		})
	})
}

func (r SchemaResource) basic(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_schema" "test" {
			database 	= "%[2]s"
			name     	= "tfschema_%[3]s"
		}
		`, template, connection, name)
}

func (r SchemaResource) withOwner(connection string, name string) string {
	return fmt.Sprintf(
		`
		resource "azuresql_role" "owner" {
			database 		= "%[2]s"
			name           	= "tfschema_owner_%[3]s"
		}

		resource "azuresql_schema" "test" {
			database 	= "%[2]s"
			name       	= "tfschema_%[3]s"
			owner		= azuresql_role.owner.id
		}
		`, r.template(), connection, name)
}

func (r SchemaResource) updateOwner(connection string, name string, owner int) string {
	return fmt.Sprintf(
		`
		resource "azuresql_role" "owner1" {
			database 		= "%[2]s"
			name           	= "tfschema_owner1_%[3]s"
		}

		resource "azuresql_role" "owner2" {
			database 		= "%[2]s"
			name          	= "tfschema_owner2_%[3]s"
		}

		resource "azuresql_schema" "test" {
			database 	= "%[2]s"
			name       	= "tfschema_%[3]s"
			owner		= azuresql_role.owner%[4]d.id
		}
		`, r.template(), connection, name, owner)
}

func (r SchemaResource) duplicate_schema(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_schema" "test" {
			database 	= "%[2]s"
			name     	= "tfschema_%[3]s"
		}

		resource "azuresql_schema" "test2" {
			database 	= "%[2]s"
			name     	= "tfschema_%[3]s"
		}
		`, template, connection, name)
}

func (r SchemaResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
