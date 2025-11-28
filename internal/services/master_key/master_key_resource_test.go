package master_key_test

import (
	"fmt"
	"testing"

	"terraform-provider-azuresql/internal/acceptance"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type MasterKeyResource struct{}

func TestAccCreateMasterKey(t *testing.T) {
	r := MasterKeyResource{}
	kinds := []acceptance.BackendKind{
		acceptance.BackendFabric,
		acceptance.BackendSQLServer,
		acceptance.BackendSynapseServerless,
		acceptance.BackendSynapseDedicated,
	}

	acceptance.ForEachBackend(t, kinds, func(t *testing.T, backend acceptance.Backend, data acceptance.TestData) {
		connection := acceptance.DatabaseConnection(t, backend, data)
		resource.Test(t, resource.TestCase{
			ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
			Steps: []resource.TestStep{
				{
					Config: r.basic(connection),
				},
			},
		})
	})
}

func (r MasterKeyResource) basic(connection string) string {
	return fmt.Sprintf(`
	%[1]s

	resource "azuresql_master_key" "test" {
		database 	= "%[2]s"
	}
`, r.template(), connection)
}

func (r MasterKeyResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
