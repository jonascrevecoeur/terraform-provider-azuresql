package synapseserver_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDedicatedDataSource(t *testing.T) {
	acceptance.PreCheck(t)
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   basicDedicated("abc"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.azuresql_synapseserver_dedicated.test", "id", "synapsededicated::abc:1433"),
				),
			},
		},
	})
}

func TestAccDedicatedDataSourcePort(t *testing.T) {
	acceptance.PreCheck(t)
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   basicDedicatedPort("abc", 15),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.azuresql_synapseserver_dedicated.test", "id", "synapsededicated::abc:15"),
				),
			},
		},
	})
}

func basicDedicated(name string) string {
	template := templateDedicated()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_synapseserver_dedicated" "test" {
			name = "%[2]s"
		}
		`, template, name)
}

func basicDedicatedPort(name string, port int64) string {
	template := templateDedicated()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_synapseserver_dedicated" "test" {
			name = "%[2]s"
			port = "%[3]d"
		}
		`, template, name, port)
}

func templateDedicated() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
