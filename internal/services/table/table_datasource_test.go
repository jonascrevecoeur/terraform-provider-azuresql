package table_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type tableDatasource struct{}

func TestAccReadTable(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := tableDatasource{}

	connections := []string{
		data.SQLDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))

		acceptance.ExecuteSQL(connection, fmt.Sprintf("create table dbo.tftable_%s (col1 int)", data.RandomString))
		defer acceptance.ExecuteSQL(connection, fmt.Sprintf("DROP table dbo.tftable_%s", data.RandomString))

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

func (r tableDatasource) basic(connection string, name string) string {
	return fmt.Sprintf(
		`
		provider "azuresql" {
		}

		data "azuresql_schema" "dbo" {
			database 	= "%[1]s"
			name 		= "dbo"
		}

		data "azuresql_table" "test" {
			database  	= "%[1]s"
			schema		= data.azuresql_schema.dbo.id
			name    	= "tftable_%[2]s"
		}
		`, connection, name)
}
