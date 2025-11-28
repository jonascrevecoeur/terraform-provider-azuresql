package view_test

import (
	"fmt"
	"testing"

	"terraform-provider-azuresql/internal/acceptance"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ViewResource struct{}

func TestAccCreateViewBasic(t *testing.T) {
	r := ViewResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.DatabaseConnection(t, backend, data)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.basic(connection, data.RandomString),
				},
				{
					Config:            r.basic(connection, data.RandomString),
					ResourceName:      "azuresql_view.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccCreateViewWithOptions(t *testing.T) {
	r := ViewResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSQLServer,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.DatabaseConnection(t, backend, data)

		acceptance.ExecuteSQL(connection, fmt.Sprintf("create table dbo.tftable_%s (col1 int)", data.RandomString))
		defer acceptance.ExecuteSQL(connection, fmt.Sprintf("DROP table dbo.tftable_%s", data.RandomString))

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.with_options(connection, data.RandomString),
				},
				{
					Config:            r.with_options(connection, data.RandomString),
					ResourceName:      "azuresql_view.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccCreateViewWithEOT(t *testing.T) {
	r := ViewResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.DatabaseConnection(t, backend, data)

		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.with_eot(connection, data.RandomString),
				},
			},
		})
	})
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

func (r ViewResource) with_options(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_schema" "dbo" {
			database 	= "%[2]s"
			name 		= "dbo"
		}

		resource "azuresql_view" "test" {
			database 		= "%[2]s"
			name        	= "tfview_%[3]s"
			schema			= data.azuresql_schema.dbo.id
			schemabinding	= true
			check_option	= true
			definition		= "select top 10 col1 from dbo.tftable_%[3]s"
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
