package schema_test

import (
	"fmt"
	"regexp"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SchemaResource struct{}

func TestAccCreateSchemaBasic(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SchemaResource{}

	connections := []string{
		//data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
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
					ResourceName:             "azuresql_schema.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

func TestAccCreateDuplicateResource(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SchemaResource{}

	connections := []string{
		data.SQLDatabase_connection,
		//data.SynapseDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.duplicate_schema(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					ExpectError:              regexp.MustCompile("You can import this resource using"),
				},
			},
		})
	}
}

func TestAccCreateSchemaWithOwner(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SchemaResource{}

	connections := []string{
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.withOwner(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func TestAccUpdateSchemaOwner(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SchemaResource{}

	connections := []string{
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.updateOwner(connection, data.RandomString, 1),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"azuresql_schema.test", "owner",
							"azuresql_role.owner1", "id"),
					),
				},
				{
					Config:                   r.updateOwner(connection, data.RandomString, 2),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"azuresql_schema.test", "owner",
							"azuresql_role.owner2", "id"),
					),
				},
			},
		})
	}
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
