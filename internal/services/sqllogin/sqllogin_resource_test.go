package login_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SQLLoginResource struct{}

func TestAccCreateLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SQLLoginResource{}

	connections := []string{
		data.SQLServer_connection,
		data.SynapseServer_connection,
	}

	for _, connection := range connections {
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.basic(connection, "tftest_"+data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("azuresql_login.test", "name", "tftest_"+data.RandomString),
					),
				},
			},
		})
	}
}

func TestAccSynapseServerCreateLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SQLLoginResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   r.basic(data.SynapseServer_connection, "tftest_"+data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_login.test", "name", "tftest_"+data.RandomString),
				),
			},
		},
	})
}

func (r SQLLoginResource) basic(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_login" "test" {
			server  = "%[2]s"
			name    = "%[3]s"
		}
		`, template, connection, name)
}

func (r SQLLoginResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
