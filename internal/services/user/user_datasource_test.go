package user_test

import (
	"fmt"
	"os"
	"regexp"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type UserDataSource struct{}

func TestAccSQLReadADGroup(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserDataSource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.basic_server(data.SQLServer_connection, os.Getenv("AZURE_AD_USER"), "AzureAD"),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.azuresql_user.test", "type", "AD group"),
				),
			},
		},
	})
}

func TestAccSynapseServerReadUserWithLogin(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserDataSource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// use a dynamic configuration with the random name from above
				Config:                   r.database_with_login(data.SQLServer_connection, data.SQLDatabase_connection, data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.azuresql_user.test", "type", "SQL user"),
				),
			},
		},
	})
}

func TestAccSQLReadNonExistingUser(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserDataSource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.server_user_doesnt_exist(data.SQLServer_connection, data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				ExpectError:              regexp.MustCompile("User .* not found on connection"),
			},
		},
	})
}

func TestAccSQLReadInvalidServer(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := UserDataSource{}
	resource.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:                   r.server_user_doesnt_exist(data.RandomString, data.RandomString),
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				ExpectError:              regexp.MustCompile("connection id .* is invalid"),
			},
		},
	})
}

func (r UserDataSource) basic_server(connection string, username string, authentication string) string {
	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_user" "test" {
			server 		= "%[2]s"
			name      	= "%[3]s"
			depends_on 	= [azuresql_user.test]
		}
		`, UserResource{}.basic_server(connection, username, authentication), connection, username)
}

func (r UserDataSource) database_with_login(server string, database string, random string) string {
	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_user" "test" {
			database        = "%[3]s"
			name           	= "user_%[2]s"
			depends_on 		= [azuresql_user.test]
		}

		`, UserResource{}.database_with_login(server, database, random), random, database)
}

func (r UserDataSource) server_user_doesnt_exist(server string, random string) string {
	return fmt.Sprintf(
		`
		provider "azuresql" {
		}
		
		data "azuresql_user" "test" {
			server        = "%[1]s"
			name           	= "user_%[2]s"
		}

		`, server, random)
}
