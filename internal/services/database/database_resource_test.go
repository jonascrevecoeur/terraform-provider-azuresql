package database_test

import (
	"fmt"
	"testing"

	"terraform-provider-azuresql/internal/acceptance"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type DatabaseResource struct{}

func TestAccCreateDatabase(t *testing.T) {
	r := DatabaseResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.ServerConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.basic(connection, data.RandomString),
				},
				{
					Config:            r.basic(connection, data.RandomString),
					ResourceName:      "azuresql_database.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccRenameDatabase(t *testing.T) {
	r := DatabaseResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.ServerConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.basic(connection, data.RandomString),
				},
				{
					Config: r.basic(connection, data.RandomString+"2"),
				},
			},
		})
	})
}

func TestAccCreateSchemaInNewDatabase(t *testing.T) {
	r := DatabaseResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.ServerConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.create_schema_in_new_database(connection, data.RandomString),
				},
			},
		})
	})
}

func (r DatabaseResource) basic(server string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_database" "test" {
			server	 	= "%[2]s"
			name     	= "tfdatabase_%[3]s"
		}
		`, template, server, name)
}

func (r DatabaseResource) create_schema_in_new_database(server string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_database" "test" {
			server	 	= "%[2]s"
			name     	= "tfdatabase_%[3]s"
		}

		resource "azuresql_schema" "test" {
			database	= azuresql_database.test.id
			name     	= "tfschema_%[3]s"
		}

		`, template, server, name)
}

func (r DatabaseResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
