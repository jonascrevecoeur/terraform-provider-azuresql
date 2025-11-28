package role_test

import (
	"fmt"
	"testing"
	"time"

	"terraform-provider-azuresql/internal/acceptance"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

type RoleResource struct{}

func TestAccCreateRoleBasic(t *testing.T) {
	r := RoleResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		for _, connection := range roleConnections(backend, data) {
			connection := connection
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: r.basic(connection, data.RandomString),
					},
					{
						Config:            r.basic(connection, data.RandomString),
						ResourceName:      "azuresql_role.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		}
	})
}

func TestAccCreateRoleWithOwner(t *testing.T) {
	r := RoleResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		for _, connection := range roleConnections(backend, data) {
			connection := connection
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: r.withOwner(connection, data.RandomString),
					},
				},
			})
		}
	})
}

func TestAccUpdateRoleName(t *testing.T) {
	r := RoleResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		for _, connection := range roleConnections(backend, data) {
			connection := connection
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: r.basic(connection, data.RandomString),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("azuresql_role.test", "name", "tfrole_"+data.RandomString),
						),
					},
					{
						Config: r.basic(connection, "updated"+data.RandomString),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("azuresql_role.test", "name", "tfrole_updated"+data.RandomString),
						),
					},
				},
			})
		}
	})
}

func delay_next_step(d time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		time.Sleep(d)

		return nil
	}
}

func TestAccUpdateRoleOwner(t *testing.T) {
	r := RoleResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		for _, connection := range roleConnections(backend, data) {
			connection := connection
			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: r.updateOwner(connection, data.RandomString, 1),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttrPair("azuresql_role.test", "owner", "azuresql_role.owner1", "id"),
							resource.TestCheckResourceAttr("azuresql_role.test", "name", "tfrole_"+data.RandomString),
							delay_next_step(30*time.Second),
						),
					},
					{
						Config: r.updateOwner(connection, data.RandomString, 2),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttrPair("azuresql_role.test", "owner", "azuresql_role.owner2", "id"),
						),
					},
				},
			})
		}
	})
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

func roleConnections(backend acceptance.Backend, data acceptance.TestData) []string {
	var connections []string
	if backend.Kind == acceptance.BackendSQLServer || backend.Kind == acceptance.BackendFabric {
		if server := backend.ServerConn(data); server != "" {
			connections = append(connections, server)
		}
	}

	if database := backend.DatabaseConn(data); database != "" {
		connections = append(connections, database)
	}

	return connections
}
