package login_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSQLServerCreateLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   basic(data.SQLServer_connection, "tftest_"+data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_login.test", "name", "tftest_"+data.RandomString),
				),
			},
		},
	})
}

func TestAccSynapseServerCreateLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   basic(data.SynapseServer_connection, "tftest_"+data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_login.test", "name", "tftest_"+data.RandomString),
				),
			},
		},
	})
}

func basic(connection string, name string) string {
	template := template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_login" "test" {
			server  = "%[2]s"
			name    = "%[3]s"
		}
		`, template, connection, name)
}

func template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
