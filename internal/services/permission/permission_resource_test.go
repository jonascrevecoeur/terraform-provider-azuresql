package permission_test

import (
	"fmt"
	"strings"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type PermissionResource struct{}

func TestAccCreatePermissionSchemaRole(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := PermissionResource{}

	connections := []string{
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.schemaRole(connection, data.RandomString, []string{"select", "delete"}),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func (r PermissionResource) schemaRole(connection string, name string, permissions []string) string {

	return fmt.Sprintf(`
		%[1]s

		locals {
			permissions = ["select", "delete"]
		}

		resource "azuresql_schema" "test" {
			database 	= "%[2]s"
			name     	= "tfschema_%[3]s"
		}

		resource "azuresql_role" "test" {
			database 	= "%[2]s"
			name        = "tfrole_%[3]s"
		}

		resource "azuresql_permission" "test" {
			count       = length(local.permissions)
			database 	= "%[2]s"
			scope 		= azuresql_schema.test.id
			principal   = azuresql_role.test.id
			permission  = local.permissions[count.index]
		}
	`, r.template(), connection, name, strings.Join(permissions, "\",\""))
}

func (r PermissionResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
