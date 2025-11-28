package login_test

import (
	"fmt"
	"regexp"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SQLLoginDatasource struct{}

func TestAccSQLServerReadLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SQLLoginDatasource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   r.basic(data.SQLServer_connection, "tftest_"+data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.azuresql_login.test", "name", "tftest_"+data.RandomString),
				),
			},
		},
	})
}

func TestAccSynapseServerReadLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SQLLoginDatasource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   r.basic(data.SynapseServer_connection, "tftest_"+data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.azuresql_login.test", "name", "tftest_"+data.RandomString),
				),
			},
		},
	})
}

func TestAccSQLServerReadNonExistentLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SQLLoginDatasource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   r.read_non_existent(data.SQLServer_connection, "tftest_"+data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				ExpectError:              regexp.MustCompile("Login [a-z0-9_]+ not found"),
			},
		},
	})
}

func (SQLLoginDatasource) basic(connection string, name string) string {
	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_login" "test" {
			server  = "%[2]s"
			name    = "%[3]s"
			depends_on = [azuresql_login.test]
		}
		`, SQLLoginResource{}.basic(connection, name), connection, name)
}

func (r SQLLoginDatasource) read_non_existent(connection string, name string) string {
	return fmt.Sprintf(
		`
		data "azuresql_login" "test" {
			server  = "%[1]s"
			name    = "%[2]s"
		}
		`, connection, name)
}
