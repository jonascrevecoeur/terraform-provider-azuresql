package securitypredicate_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SecurityPredicateResource struct{}

func TestAccCreateSecurityPredicateBasic(t *testing.T) {
	acceptance.PreCheck(t)

	data := acceptance.BuildTestData(t)
	r := SecurityPredicateResource{}

	connections := []string{
		data.FabricDatabase_connection,
		data.SQLDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))

		acceptance.ExecuteSQL(connection, fmt.Sprintf("create table dbo.tftable_%s (col1 int)", data.RandomString))
		defer acceptance.ExecuteSQL(connection, fmt.Sprintf("DROP table dbo.tftable_%s", data.RandomString))

		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.basic(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
				{
					Config:                   r.basic(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					ResourceName:             "azuresql_security_predicate.test",
					ImportState:              true,
					// no verification as the rule specified in the database might differ (additional []())
					//ImportStateVerify:        true,
				},
			},
		})
	}
}

func (r SecurityPredicateResource) basic(connection string, name string) string {
	template := r.template(connection, name)

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_security_predicate" "test" {
			database 			= "%[2]s"
			security_policy   	= azuresql_security_policy.test.id
			table			   	= data.azuresql_table.test.id
			rule				= "[dbo].[tffunction_%[3]s]([col1])"
			type				= "filter"
			depends_on          = [azuresql_function.test]
		}
		`, template, connection, name)
}

func (r SecurityPredicateResource) template(connection string, name string) string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}

		data "azuresql_schema" "dbo" {
			database 	= "%[1]s"
			name 		= "dbo"
		}

		data "azuresql_table" "test" {
			database 	= "%[1]s"
			name 		= "tftable_%[2]s"
		}

		resource "azuresql_function" "test" {
			database 	= "%[1]s"
			name        = "tffunction_%[2]s"
			schema		= data.azuresql_schema.dbo.id
			raw         = <<-EOT
				create function dbo.tffunction_%[2]s(@test int)
				returns table 
				with SCHEMABINDING AS
				return  
				select 1 as a
			EOT
		}

		resource "azuresql_security_policy" "test" {
			database 	= "%[1]s"
			name        = "tfsecurity_policy_%[2]s"
			schema		= data.azuresql_schema.dbo.id
			depends_on  = [azuresql_function.test]
		}
	`, connection, name)
}
