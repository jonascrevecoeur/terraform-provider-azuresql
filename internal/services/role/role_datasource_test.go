package role_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type roleDataSource struct{}

func TestAccReadRole(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := roleDataSource{}
	connections := []string{
		data.SQLServer_connection,
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
		data.SynapseDedicatedDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.basic(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func (r roleDataSource) basic(connection string, name string) string {
	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_role" "test" {
			%[2]s
			name        = "tfrole_%[3]s"
			depends_on 	= [azuresql_role.test]
		}

		`, RoleResource{}.basic(connection, name), acceptance.TerraformConnectionId(connection), name)
}
