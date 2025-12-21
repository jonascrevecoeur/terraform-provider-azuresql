package master_key_test

import (
	"fmt"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type MasterKeyResource struct{}

func TestAccCreateMasterKey(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := MasterKeyResource{}

	connections := []string{
		data.FabricDatabase_connection,
		data.SQLDatabase_connection,
		data.SynapseDatabase_connection,
		data.SynapseDedicatedDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.basic(connection),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
				},
			},
		})
	}
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
