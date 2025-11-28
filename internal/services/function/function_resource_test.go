package function_test

import (
	"fmt"
	"testing"

	"terraform-provider-azuresql/internal/acceptance"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type FunctionResource struct{}

func TestAccCreateFunctionBasic(t *testing.T) {
	r := FunctionResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
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
					ResourceName:      "azuresql_function.test",
					ImportState:       true,
					ImportStateVerify: true,
				},
			},
		})
	})
}

func TestAccCreateFunctionProps(t *testing.T) {
	r := FunctionResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendSQLServer,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.DatabaseConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.propsapi(connection, data.RandomString),
				},
			},
		})
	})
}

func (r FunctionResource) basic(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_schema" "dbo" {
			database 	= "%[2]s"
			name 		= "dbo"
		}

		resource "azuresql_function" "test" {
			database 	= "%[2]s"
			name        = "tffunction_%[3]s"
			schema		= data.azuresql_schema.dbo.id
			raw         = <<-EOT
				create FUNCTION dbo.tffunction_%[3]s()
				returns table 
				with SCHEMABINDING AS
				return  
				select 1 as a
			EOT
		}
		`, template, connection, name)
}

func (r FunctionResource) propsapi(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_schema" "dbo" {
			database 	= "%[2]s"
			name 		= "dbo"
		}

		resource "azuresql_function" "test" {
			database 	= "%[2]s"
			name        = "tffunction_%[3]s"
			schema		= data.azuresql_schema.dbo.id
			properties = {
				arguments = [
				  {
					name = "a"
					type = "int"
				  },
				  {
					name = "b"
					type = "int"
				  }
				]
				executor      = "self"
				return_type   = "int"
				schemabinding = true
				definition    = "@a + @b"
			  }
		}
		`, template, connection, name)
}

func (r FunctionResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
