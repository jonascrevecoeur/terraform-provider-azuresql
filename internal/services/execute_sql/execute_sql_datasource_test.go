package execute_sql_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type executeSQLDataSource struct{}

func TestAccCreateTable(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := executeSQLDataSource{}
	connections := []string{
		data.FabricDatabase_connection,
		data.SQLDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))

		defer acceptance.ExecuteSQL(connection, fmt.Sprintf("DROP table dbo.test_%s", data.RandomString))

		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.create_table(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func TestAccCreateMultipleTables(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := executeSQLDataSource{}
	connections := []string{
		data.FabricDatabase_connection,
		data.SQLDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))

		defer acceptance.ExecuteSQL(connection, fmt.Sprintf("DROP table dbo.test1_%s", data.RandomString))
		defer acceptance.ExecuteSQL(connection, fmt.Sprintf("DROP table dbo.test2_%s", data.RandomString))

		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.create_tables(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func TestAccExecuteSQLCreateLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := executeSQLDataSource{}
	connections := []string{
		data.SQLServer_connection,
		data.SynapseServer_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))

		defer acceptance.ExecuteSQL(connection, fmt.Sprintf("DROP login test_%s", data.RandomString))

		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.create_login(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func (r executeSQLDataSource) create_table(connection string, name string) string {
	return fmt.Sprintf(
		`
		provider "azuresql" {
		}

		data "azuresql_execute_sql" "test" {
			database    = "%[1]s"
			sql 		= <<-EOT
				IF OBJECT_ID('test_%[2]s', 'U') IS NULL
				create table test_%[2]s(
					col1 float
				)
			EOT
		}
		`, connection, name)
}

func (r executeSQLDataSource) create_tables(connection string, name string) string {
	return fmt.Sprintf(
		`
		provider "azuresql" {
		}

		data "azuresql_execute_sql" "test1" {
			database    = "%[1]s"
			sql 		= <<-EOT
				IF OBJECT_ID('test1_%[2]s', 'U') IS NULL
				create table test1_%[2]s(
					col1 float
				)

				IF OBJECT_ID('test2_%[2]s', 'U') IS NULL
				create table test2_%[2]s(
					col1 float
				)
			EOT
		}

		`, connection, name)
}

func (r executeSQLDataSource) create_login(connection string, name string) string {
	return fmt.Sprintf(
		`
		provider "azuresql" {
		}

		data "azuresql_execute_sql" "test" {
			server		= "%[1]s"
			sql 		= <<-EOT
				IF NOT EXISTS 
					(SELECT name from master.sys.sql_logins where name = 'test_%[2]s')
				BEGIN
					create login test_%[2]s with password = '1e5gegeg15jijf!bg!!'
				END
			EOT
		}
		`, connection, name)
}
