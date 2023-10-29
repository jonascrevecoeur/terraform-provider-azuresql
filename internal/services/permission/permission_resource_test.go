package permission_test

import (
	"fmt"
	"strings"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type PermissionResource struct{}

func TestAccCreatePermissionDatabaseRole(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := PermissionResource{}

	connections := []string{
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.databaseRole(connection, data.RandomString, "create table"),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
				{
					Config:                   r.databaseRole(connection, data.RandomString, "create table"),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					ResourceName:             "azuresql_permission.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

/*func TestAccCreatePermissionServerRole(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := PermissionResource{}

	connections := []string{
		data.SQLServer_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.serverRole(connection, data.RandomString, "alter any event notification"),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}*/

func TestAccCreatePermissionSchemaRole(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := PermissionResource{}

	connections := []string{
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))

		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.schemaRole(connection, data.RandomString, []string{"select", "delete"}),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: testAccCheckPermissionId(
						"azuresql_permission.test.0", "azuresql_schema.test", "schema_id",
						"azuresql_role.test", "schema", "select",
					),
				},
			},
		})
	}
}

func TestAccCreatePermissionDatabaseScopedCredentialUser(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := PermissionResource{}

	connections := []string{
		data.SQLDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))

		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.databaseScopedCredential_user(connection, data.RandomString, "control"),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func TestAccCreatePermissionTableRole(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := PermissionResource{}

	connections := []string{
		data.SQLDatabase_connection,
		//data.SynapseDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))

		acceptance.ExecuteSQL(connection, fmt.Sprintf("create table dbo.tftable_%s (col1 int)", data.RandomString))
		defer acceptance.ExecuteSQL(connection, fmt.Sprintf("DROP table dbo.tftable_%s", data.RandomString))

		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.tableRole(connection, data.RandomString, []string{"select"}),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: testAccCheckPermissionId(
						"azuresql_permission.test.0", "data.azuresql_table.test", "object_id",
						"azuresql_role.test", "object", "select",
					),
				},
			},
		})
	}
}

func testAccCheckPermissionId(permission_obj string, target_obj string, target_field string, principal_obj string, permissionType string, permissionString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		permission, ok := s.RootModule().Resources[permission_obj]
		permission_id := permission.Primary.ID

		if !ok {
			return fmt.Errorf("Not found: %s", permission_obj)
		}

		target, ok := s.RootModule().Resources[target_obj]
		if !ok {
			return fmt.Errorf("Not found: %s", target_obj)
		}
		targetId := target.Primary.Attributes[target_field]

		principal, ok := s.RootModule().Resources[principal_obj]
		if !ok {
			return fmt.Errorf("Not found: %s", principal_obj)
		}
		principalId := principal.Primary.Attributes["principal_id"]

		if !strings.Contains(permission_id,
			fmt.Sprintf("permission/%s/%s/%s/%s", principalId, permissionString, permissionType, targetId)) {
			return fmt.Errorf("permission_id %s doesn't adhere to the required schema permission/%s/%s/%s/%s",
				permission_id, principalId, permissionString, permissionType, targetId)
		}

		return nil
	}
}

func (r PermissionResource) databaseRole(connection string, name string, permission string) string {
	return fmt.Sprintf(`
	%[1]s

	resource "azuresql_role" "test" {
		database 	= "%[2]s"
		name        = "tfrole_%[3]s"
	}

	resource "azuresql_permission" "test" {
		database 	= "%[2]s"
		scope 		= "%[2]s"
		principal   = azuresql_role.test.id
		permission  = "%[4]s"
	}
`, r.template(), connection, name, permission)
}

func (r PermissionResource) serverRole(connection string, name string, permission string) string {
	return fmt.Sprintf(`
	%[1]s

	resource "azuresql_role" "test" {
		server 		= "%[2]s"
		name        = "tfrole_%[3]s"
	}

	resource "azuresql_permission" "test" {
		server 		= "%[2]s"
		scope 		= "%[2]s"
		principal   = azuresql_role.test.id
		permission  = "%[4]s"
	}
`, r.template(), connection, name, permission)
}

func (r PermissionResource) schemaRole(connection string, name string, permissions []string) string {

	return fmt.Sprintf(`
		%[1]s

		locals {
			permissions = ["%[4]s"]
		}

		resource "azuresql_schema" "test" {
			database 	= "%[2]s"
			name     	= "tfschema_%[3]s"
		}

		resource "azuresql_role" "test" {
			database 	= "%[2]s"
			name        = "tfrole_%[3]s"
		}

		resource "azuresql_permission" "test" {
			count       = length(local.permissions)
			database 	= "%[2]s"
			scope 		= azuresql_schema.test.id
			principal   = azuresql_role.test.id
			permission  = local.permissions[count.index]
		}
	`, r.template(), connection, name, strings.Join(permissions, "\",\""))
}

func (r PermissionResource) tableRole(connection string, name string, permissions []string) string {

	return fmt.Sprintf(`
		%[1]s

		locals {
			permissions = ["%[4]s"]
		}

		data "azuresql_table" "test" {
			database 	= "%[2]s"
			name     	= "tftable_%[3]s"
		}

		resource "azuresql_role" "test" {
			database 	= "%[2]s"
			name        = "tfrole_%[3]s"
		}

		resource "azuresql_permission" "test" {
			count       = length(local.permissions)
			database 	= "%[2]s"
			scope 		= data.azuresql_table.test.id
			principal   = azuresql_role.test.id
			permission  = local.permissions[count.index]
		}
	`, r.template(), connection, name, strings.Join(permissions, "\",\""))
}

func (r PermissionResource) databaseScopedCredential_user(connection string, name string, permission string) string {
	return fmt.Sprintf(`
	%[1]s

	resource "azuresql_user" "test" {
		database 		= "%[2]s"
		name        	= "tfuser_%[3]s"
		authentication 	= "WithoutLogin"
	}

	resource "azuresql_master_key" "test" {
		database 		= "%[2]s"
	}

	resource "azuresql_database_scoped_credential" "test" {
		database 		= "%[2]s"
		name			= "tfdsc_%[3]s"
		identity		= "test"

		depends_on = [azuresql_master_key.test]
	}

	resource "azuresql_permission" "test" {
		database 	= "%[2]s"
		scope 		= azuresql_database_scoped_credential.test.id
		principal   = azuresql_user.test.id
		permission  = "%[4]s"
	}
`, r.template(), connection, name, permission)
}

func (r PermissionResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
