package role_assignment_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type RoleAssignmentResource struct{}

func TestAccAssignRoletoUser(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := RoleAssignmentResource{}

	connections := []string{
		data.SQLServer_connection,
		data.SQLDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.assignRoleToUser(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
				{
					Config:                   r.assignRoleToUser(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					ResourceName:             "azuresql_role_assignment.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

func (r RoleAssignmentResource) assignRoleToUser(connection string, name string) string {

	return fmt.Sprintf(`
	%[1]s

	resource "azuresql_user" "user" {
		%[2]s
		name        	= "tfuser_%[3]s"
		authentication 	= "WithoutLogin"
	}

	resource "azuresql_role" "test" {
		%[2]s
		name        = "tfrole_%[3]s"
	}

	resource "azuresql_role_assignment" "test" {
		%[2]s
		role 		= azuresql_role.test.id
		principal   = azuresql_user.user.id
	}
`, r.template(), acceptance.TerraformConnectionId(connection), name)
}

func (r RoleAssignmentResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
