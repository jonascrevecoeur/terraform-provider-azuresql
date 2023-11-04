package view_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ViewResource struct{}

func TestAccCreateViewBasic(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := ViewResource{}

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
				{
					Config:                   r.basic(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					ResourceName:             "azuresql_view.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

func TestAccCreateViewWithEOT(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := ViewResource{}

	connections := []string{
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))

		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.with_eot(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func (r ViewResource) basic(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_schema" "dbo" {
			database 	= "%[2]s"
			name 		= "dbo"
		}

		resource "azuresql_view" "test" {
			database 	= "%[2]s"
			name        = "tfview_%[3]s"
			schema		= data.azuresql_schema.dbo.id
			definition	= "select top 10 * from sys.objects"
		}
		`, template, connection, name)
}

func (r ViewResource) with_eot(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_schema" "dbo" {
			database 	= "%[2]s"
			name 		= "dbo"
		}

		resource "azuresql_view" "test" {
			database 	= "%[2]s"
			name        = "tfview_%[3]s"
			schema		= data.azuresql_schema.dbo.id
			definition	= <<-EOT
				select top 10 * from sys.objects
			EOT
		}
		`, template, connection, name)
}

func (r ViewResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
