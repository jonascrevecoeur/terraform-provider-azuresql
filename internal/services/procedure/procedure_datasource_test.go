package procedure_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type procedureDataSource struct{}

func TestAccReadProcedure(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := procedureDataSource{}
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

func (r procedureDataSource) basic(connection string, name string) string {
	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_procedure" "test" {
			database 	= "%[2]s"
			name        = "tfprocedure_%[3]s"
			schema		= data.azuresql_schema.dbo.id
			depends_on  = [azuresql_procedure.test]
		}

		`, ProcedureResource{}.basic(connection, name), connection, name)
}
