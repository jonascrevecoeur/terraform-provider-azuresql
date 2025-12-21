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
			{
				Config:                   r.basic_server(data.SQLServer_connection, data.RandomString, "WithoutLogin"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				ResourceName:             "azuresql_user.test",
				ImportState:              true,
				ImportStateVerify:        true,
			},
		},
	})
}

func TestAccUseEnvironmentCredentials(t *testing.T) {
	t.Setenv("AZURE_CLIENT_SECRET", os.Getenv("AZURE_CLIENT_SECRET_OPT"))
	t.Setenv("AZURE_CLIENT_ID", os.Getenv("AZURE_CLIENT_ID_OPT"))
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.basic_database(data.SQLDatabase_connection, data.RandomString, "WithoutLogin"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
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

func TestAccFabricServerCreateADGroup(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.basic_server(data.FabricServer_connection, os.Getenv("AZURE_AD_GROUP"), "AzureAD"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("azuresql_user.test", "type", "AD group"),
				),
			},
		},
	})
}

func TestAccFabricDatabaseCreateADGroup(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.basic_database(data.FabricDatabase_connection, os.Getenv("AZURE_AD_GROUP"), "AzureAD"),
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

func TestAccSQLDatabaseCreateUserWithEntraIDIdentity(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: r.entraid_identity_database(
					data.SQLDatabase_connection,
					"azuresql-sid",
					"11111111-1111-1111-1111-111111111111"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			},
		},
	})
}

func TestAccSynapseServerCreateADGroup(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	for _, connection := range []string{data.SynapseServer_connection, data.SynapseDedicatedServer_connection} {
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.basic_server(connection, os.Getenv("AZURE_AD_GROUP"), "AzureAD"),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					ExpectError:              regexp.MustCompile("In Synapse users cannot be created at server level"),
				},
			},
		})
	}
}

func TestAccSynapseDatabaseCreateADGroup(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}
	for _, connection := range []string{data.SynapseDatabase_connection, data.SynapseDedicatedDatabase_connection} {
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.basic_database(connection, os.Getenv("AZURE_AD_GROUP"), "AzureAD"),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("azuresql_user.test", "type", "AD group"),
					),
				},
			},
		})
	}
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

func TestAccDBSQLLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserResource{}

	connections := []string{
		data.SQLDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.user_with_database_login(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
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

func (r UserResource) entraid_identity_database(connection string, username string, entraid_identifier string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_user" "test" {
			database  	   		= "%[2]s"
			name           		= "%[3]s"
			entraid_identifier 	= "%[4]s"
			authentication 		= "AzureAD"
		}
		`, template, connection, username, entraid_identifier)
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

func (r UserResource) user_with_database_login(connection string, name string) string {
	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_user" "test" {
			database 		= "%[2]s"
			name           	= "user_database_login_%[3]s"
			authentication 	= "DBSQLLogin"
			password 		= "Difficultpassword12!abc13!!"
		}
		`, r.template(), connection, name)
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
			check_database_exists = false
		}
	`)
}
