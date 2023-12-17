package database_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type DatabaseResource struct{}

func TestAccCreateDatabase(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := DatabaseResource{}

	connections := []string{
		data.SynapseServer_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.basic(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
				{
					Config:                   r.basic(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					ResourceName:             "azuresql_database.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

func TestAccCreateSchemaInNewDatabase(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := DatabaseResource{}

	connections := []string{
		data.SynapseServer_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.create_schema_in_new_database(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
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
