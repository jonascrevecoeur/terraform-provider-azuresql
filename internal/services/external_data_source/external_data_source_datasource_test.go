package external_data_source_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type externalDataSourceDataSource struct{}

func TestAccReadExternalDataSource(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := externalDataSourceDataSource{}
	connections := []string{
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

func (r externalDataSourceDataSource) basic(database string, name string) string {
	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_external_data_source" "test" {
			database	= "%[2]s"
			name        = "tfdatasource_%[3]s"
			depends_on 	= [azuresql_external_data_source.test]
		}

		`, ExternalDataSourceResource{}.basic(database, name), database, name)
}
