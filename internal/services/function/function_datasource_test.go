package function_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type functionDataSource struct{}

func TestAccReadFunction(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := functionDataSource{}
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
					Config:                   r.basic(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func (r functionDataSource) basic(connection string, name string) string {
	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_function" "test" {
			database 	= "%[2]s"
			name        = "tffunction_%[3]s"
			schema		= data.azuresql_schema.dbo.id
			depends_on  = [azuresql_function.test]
		}

		`, FunctionResource{}.basic(connection, name), connection, name)
}
