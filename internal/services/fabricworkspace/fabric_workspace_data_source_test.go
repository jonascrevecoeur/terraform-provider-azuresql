package fabricworkspace_test

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
				Config:                   basic("abc"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				// compose a basic test, checking both remote and local values
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.azuresql_fabricworkspace.test", "id", "fabric::abc:1443"),
				),
			},
		},
	})
}

func basic(name string) string {
	template := template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_fabricworkspace" "test" {
			name = "%[2]s"
		}
		`, template, name)
}

func template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
