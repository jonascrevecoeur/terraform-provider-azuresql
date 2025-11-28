package database_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSource(t *testing.T) {
	acceptance.PreCheck(t)
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   basic("abc", "def"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				// compose a basic test, checking both remote and local values
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.azuresql_database.test", "id", "sqlserver::abc:1433:def"),
				),
			},
		},
	})
}

func basic(server string, database string) string {
	template := template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_sqlserver" "test" {
			name = "%[2]s"
		}

		data "azuresql_database" "test" {
			server = data.azuresql_sqlserver.test.id
			name = "%[3]s"
		}
		`, template, server, database)
}

func basic_port(name string, port int64) string {
	template := template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_sqlserver" "test" {
			name = "%[2]s"
			port = "%[3]d"
		}
		`, template, name, port)
}

func template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
