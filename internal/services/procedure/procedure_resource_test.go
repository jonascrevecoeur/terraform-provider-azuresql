package procedure_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ProcedureResource struct{}

func TestAccCreateProcedureBasic(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := ProcedureResource{}

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
					ResourceName:             "azuresql_procedure.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

func TestAccCreateProcedureProps(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := ProcedureResource{}

	connections := []string{
		data.SQLDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.propsapi(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func (r ProcedureResource) basic(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_schema" "dbo" {
			database 	= "%[2]s"
			name 		= "dbo"
		}

		resource "azuresql_procedure" "test" {
			database 	= "%[2]s"
			name        = "tfprocedure_%[3]s"
			schema		= data.azuresql_schema.dbo.id
			raw         = <<-EOT
				create procedure dbo.tfprocedure_%[3]s
				AS 
				select 1 as a
			EOT
		}
		`, template, connection, name)
}

func (r ProcedureResource) propsapi(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_schema" "dbo" {
			database 	= "%[2]s"
			name 		= "dbo"
		}

		resource "azuresql_procedure" "test" {
			database 	= "%[2]s"
			name        = "tfprocedure_%[3]s"
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
				definition    = "select @a + @b as sum"
			  }
		}
		`, template, connection, name)
}

func (r ProcedureResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
