package login_test

import (
	"fmt"
	"testing"

	"terraform-provider-azuresql/internal/acceptance"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SQLLoginResource struct{}

func TestAccCreateLogin(t *testing.T) {
	r := SQLLoginResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.ServerConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.basic(connection, "tftest_"+data.RandomString),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("azuresql_login.test", "name", "tftest_"+data.RandomString),
					),
				},
				{
					Config:            r.basic(connection, "tftest_"+data.RandomString),
					ResourceName:      "azuresql_login.test",
					ImportState:       true,
					ImportStateVerify: false,
				},
			},
		})
	})
}

func TestAccCreateLoginWithPassword(t *testing.T) {
	r := SQLLoginResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}
	password := "password12345!$"

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.ServerConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.with_password(connection, "tftest_"+data.RandomString, password),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("azuresql_login.test", "name", "tftest_"+data.RandomString),
					),
				},
				{
					Config:            r.with_password(connection, "tftest_"+data.RandomString, password),
					ResourceName:      "azuresql_login.test",
					ImportState:       true,
					ImportStateVerify: false,
				},
			},
		})
	})
}

func TestAccSynapseServerCreateLogin(t *testing.T) {
	r := SQLLoginResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.ServerConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.basic(connection, "tftest_"+data.RandomString),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("azuresql_login.test", "name", "tftest_"+data.RandomString),
					),
				},
			},
		})
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

func (r SQLLoginResource) with_password(connection string, name string, password string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_login" "test" {
			server  = "%[2]s"
			name    = "%[3]s"
			password = "%[4]s"
		}
		`, template, connection, name, password)
}

func (r SQLLoginResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
