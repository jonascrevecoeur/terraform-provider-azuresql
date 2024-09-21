package role_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type RoleResource struct{}

func TestAccCreateRoleBasic(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := RoleResource{}

	connections := []string{
		data.FabricDatabase_connection,
		data.FabricServer_connection,
		data.SQLServer_connection,
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
					ResourceName:             "azuresql_role.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

func TestAccCreateRoleWithOwner(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := RoleResource{}

	connections := []string{
		data.SQLServer_connection,
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
		data.FabricDatabase_connection,
		data.FabricServer_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.withOwner(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
}

func TestAccUpdateRoleName(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := RoleResource{}

	connections := []string{
		data.SQLServer_connection,
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
		data.FabricDatabase_connection,
		data.FabricServer_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.basic(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("azuresql_role.test", "name", "tfrole_"+data.RandomString),
					),
				},
				{
					Config:                   r.basic(connection, "updated"+data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("azuresql_role.test", "name", "tfrole_updated"+data.RandomString),
					),
				},
			},
		})
	}
}

func delay_next_step(d time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		time.Sleep(d)

		return nil
	}
}

func TestAccUpdateRoleOwner(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := RoleResource{}

	connections := []string{
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
		data.FabricDatabase_connection,
		data.FabricServer_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		print(data.RandomString)
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.updateOwner(connection, data.RandomString, 1),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"azuresql_role.test", "owner",
							"azuresql_role.owner1", "id"),
						resource.TestCheckResourceAttr("azuresql_role.test", "name", "tfrole_"+data.RandomString),
						delay_next_step(30*time.Second),
					),
				},
				{
					Config:                   r.updateOwner(connection, data.RandomString, 2),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttrPair(
							"azuresql_role.test", "owner",
							"azuresql_role.owner2", "id"),
					),
				},
			},
		})
	}
}

func (r RoleResource) basic(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		resource "azuresql_role" "test" {
			%[2]s
			name           = "tfrole_%[3]s"
		}
		`, template, acceptance.TerraformConnectionId(connection), name)
}

func (r RoleResource) withOwner(connection string, name string) string {
	return fmt.Sprintf(
		`
		resource "azuresql_role" "owner" {
			%[2]s
			name           = "tfrole_owner_%[3]s"
		}

		resource "azuresql_role" "test" {
			%[2]s
			name       	= "tfrole_%[3]s"
			owner		= azuresql_role.owner.id
		}
		`, r.template(), acceptance.TerraformConnectionId(connection), name)
}

func (r RoleResource) updateOwner(connection string, name string, owner int) string {
	return fmt.Sprintf(
		`
		resource "azuresql_role" "owner1" {
			%[2]s
			name           = "tfrole_owner1_%[3]s"
		}

		resource "azuresql_role" "owner2" {
			%[2]s
			name           = "tfrole_owner2_%[3]s"
		}

		resource "azuresql_role" "test" {
			%[2]s
			name       	= "tfrole_%[3]s"
			owner		= azuresql_role.owner%[4]d.id
		}
		`, r.template(), acceptance.TerraformConnectionId(connection), name, owner)
}

func (r RoleResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
