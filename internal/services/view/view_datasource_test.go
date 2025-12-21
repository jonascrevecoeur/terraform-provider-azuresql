package view_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type viewDatasource struct{}

func TestAccReadView(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := viewDatasource{}

	connections := []string{
		data.SQLDatabase_connection,
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
			},
		})
	}
}

func (r viewDatasource) basic(connection string, name string) string {
	return fmt.Sprintf(
		`%[1]s

		data "azuresql_view" "test" {
			database  	= "%[2]s"
			schema		= data.azuresql_schema.dbo.id
			name    	= "tfview_%[3]s"
			depends_on 	= [azuresql_view.test]
		}
		`, ViewResource{}.basic(connection, name), connection, name)
}
