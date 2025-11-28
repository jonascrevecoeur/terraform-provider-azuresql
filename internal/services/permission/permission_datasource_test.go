package permission_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type permissionDataSource struct{}

func TestAccReadPermission(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := permissionDataSource{}
	connections := []string{
		data.FabricDatabase_connection,
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
		data.SynapseDedicatedDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.schemaRole(connection, data.RandomString, []string{"select", "delete"}),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("data.azuresql_permission.test", "permissions.#", "2"),
						resource.TestCheckResourceAttr("data.azuresql_permission.test", "permissions.0", "DELETE"),
						resource.TestCheckResourceAttr("data.azuresql_permission.test", "permissions.1", "SELECT"),
					),
				},
			},
		})
	}
}

func (r permissionDataSource) schemaRole(connection string, name string, permissions []string) string {
	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_permission" "test" {
			database 	= "%[2]s"
			scope 		= azuresql_schema.test.id
			principal   = azuresql_role.test.id
			depends_on  = [azuresql_permission.test]
		}

		`, PermissionResource{}.schemaRole(connection, name, permissions), connection, name)
}
