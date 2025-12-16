package external_data_source_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type ExternalDataSourceResource struct{}

func TestAccCreateExternalDataSourceWithoutCredential(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := ExternalDataSourceResource{}

	connections := []string{
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
		data.SynapseDedicatedDatabase_connection,
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
					ResourceName:             "azuresql_external_data_source.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

func TestAccCreateExternalDataSourceWithCredential(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := ExternalDataSourceResource{}

	connections := []string{
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
		data.SynapseDedicatedDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.withCredential(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
				{
					Config:                   r.withCredential(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					ResourceName:             "azuresql_external_data_source.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

func (r ExternalDataSourceResource) basic(database string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_external_data_source" "test" {
			database	= "%[2]s"
			name		= "tfdatasource_%[3]s"
			location	= "%[3]s"
		}
		`, template, database, name)
}

func (r ExternalDataSourceResource) withCredential(database string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_master_key" "test" {
			database 	= "%[2]s"
		}

		resource "azuresql_database_scoped_credential" "test" {
			database 	= "%[2]s"
			name  		= "tfcredential_%[3]s"
			identity  	= "SHARED ACCESS SIGNATURE"
	
			depends_on = [azuresql_master_key.test]
		}

		resource "azuresql_external_data_source" "test" {
			database	= "%[2]s"
			name		= "tfdatasource_%[3]s"
			location	= "%[3]s"
			credential  = azuresql_database_scoped_credential.test.id
		}
		`, template, database, name)
}

func (r ExternalDataSourceResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
