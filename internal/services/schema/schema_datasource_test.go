package schema_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type schemaDataSource struct{}

func TestAccReadSchema(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := schemaDataSource{}
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

func (r schemaDataSource) basic(connection string, name string) string {
	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_schema" "test" {
			database    = "%[2]s"
			name        = "tfschema_%[3]s"
			depends_on 	= [azuresql_schema.test]
		}

		`, SchemaResource{}.basic(connection, name), connection, name)
}
