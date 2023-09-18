package user_test

import (
	"fmt"
	"os"
	"regexp"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type UserResource struct{}

func TestAccSQLServerCreateADUser(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   r.basic_server(data.SQLServer_connection, os.Getenv("AZURE_AD_USER"), "AzureAD"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_user.test", "type", "AD user"),
				),
			},
		},
	})
}

func TestAccSQLServerCreateUserWithoutLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   r.basic_server(data.SQLServer_connection, data.RandomString, "WithoutLogin"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_user.test", "type", "SQL user"),
				),
			},
		},
	})
}

func TestAccSQLServerCreateUserWithLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   r.server_with_login(data.SQLServer_connection, data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_user.test", "type", "SQL user"),
				),
			},
		},
	})
}

func TestAccSynapseServerCreateUserWithLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   r.database_with_login(data.SQLServer_connection, data.SQLDatabase_connection, data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_user.test", "type", "SQL user"),
				),
			},
		},
	})
}

func TestAccSQLServerCreateADGroup(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.basic_server(data.SQLServer_connection, os.Getenv("AZURE_AD_GROUP"), "AzureAD"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_user.test", "type", "AD group"),
				),
			},
		},
	})
}

func TestAccSQLDatabaseCreateADGroup(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.basic_database(data.SQLDatabase_connection, os.Getenv("AZURE_AD_GROUP"), "AzureAD"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_user.test", "type", "AD group"),
				),
			},
		},
	})
}

func TestAccSynapseServerCreateADGroup(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.basic_server(data.SynapseServer_connection, os.Getenv("AZURE_AD_GROUP"), "AzureAD"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				ExpectError:              regexp.MustCompile("In Synapse users cannot be created at server level"),
			},
		},
	})
}

func TestAccSynapseDatabaseCreateADGroup(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.basic_database(data.SynapseDatabase_connection, os.Getenv("AZURE_AD_GROUP"), "AzureAD"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_user.test", "type", "AD group"),
				),
			},
		},
	})
}

func TestAccSQLServerCannotCreateDuplicateUser(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.basic_server_duplicate(data.SQLServer_connection, os.Getenv("AZURE_AD_GROUP"), "AzureAD"),
				ExpectError:              regexp.MustCompile("already exists"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			},
		},
	})
}

func TestUseLoginFromWrongServer(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.login_server_mismatch(data.SynapseServer_connection, data.SQLServer_connection, data.RandomString),
				ExpectError:              regexp.MustCompile("Login from .* is incompatible"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			},
		},
	})
}

func (r UserResource) basic_server(connection string, username string, authentication string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_user" "test" {
			server  	   = "%[2]s"
			name           = "%[3]s"
			authentication = "%[4]s"
		}
		`, template, connection, username, authentication)
}

func (r UserResource) basic_database(connection string, username string, authentication string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_user" "test" {
			database  	   = "%[2]s"
			name           = "%[3]s"
			authentication = "%[4]s"
		}
		`, template, connection, username, authentication)
}

func (r UserResource) basic_server_duplicate(connection string, username string, authentication string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_user" "test" {
			server         = "%[2]s"
			name           = "%[3]s"
			authentication = "%[4]s"
		}

		resource "azuresql_user" "test2" {
			server         = "%[2]s"
			name           = "%[3]s"
			authentication = "%[4]s"
		}
		`, template, connection, username, authentication)
}

func (r UserResource) server_with_login(connection string, random string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_login" "test" {
			server 	= "%[2]s" 
			name 	= "login_%[3]s"
		}

		resource "azuresql_user" "test" {
			server         = "%[2]s"
			name           = "user_%[3]s"
			authentication = "SQLLogin"
			login		   = azuresql_login.test.id
		}

		`, template, connection, random)
}

func (r UserResource) database_with_login(server string, database string, random string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_login" "test" {
			server = "%[2]s"
			name   = "login_%[3]s"
		}

		resource "azuresql_user" "test" {
			database        = "%[4]s"
			name           	= "user_%[3]s"
			authentication 	= "SQLLogin"
			login		   	= azuresql_login.test.id
		}

		`, template, server, random, database)
}

func (r UserResource) login_server_mismatch(server1 string, server2 string, user string) string {
	template := r.template()
	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_login" "test" {
			server = "%[2]s"
			name   = "login_%[4]s"
		}

		resource "azuresql_user" "test" {
			server        = "%[3]s"
			name           	= "user_%[4]s"
			authentication 	= "SQLLogin"
			login		   	= azuresql_login.test.id
		}

		`, template, server1, server2, user)
}

func (r UserResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
