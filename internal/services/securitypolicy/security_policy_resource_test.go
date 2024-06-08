package securitypolicy_test

import (
	"fmt"
	"regexp"
	"terraform-provider-azuresql/internal/acceptance"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type SecurityPolicyResource struct{}

func TestAccCreateSecurityPolicyBasic(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SecurityPolicyResource{}

	connections := []string{
		data.FabricDatabase_connection,
		data.SQLDatabase_connection,
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
					ResourceName:             "azuresql_security_policy.test",
					ImportState:              true,
					ImportStateVerify:        true,
				},
			},
		})
	}
}

func TestAccCreateSecurityPolicySynapse(t *testing.T) {
	acceptance.PreCheck(t)
	data := acceptance.BuildTestData(t)
	r := SecurityPolicyResource{}

	connections := []string{
		data.SynapseDatabase_connection,
	}

	for _, connection := range connections {
		print(fmt.Sprintf("\n\nRunning test for connection %s\n\n", connection))
		resource.Test(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					Config:                   r.basic(connection, data.RandomString),
					ProtoV6ProviderFactories: acceptance.TestAccProtoV6ProviderFactories,
					ExpectError:              regexp.MustCompile("Security policies are not supported on Synapse"),
				},
			},
		})
	}
}

func (r SecurityPolicyResource) basic(connection string, name string) string {
	template := r.template()

	return fmt.Sprintf(
		`
		%[1]s

		data "azuresql_schema" "dbo" {
			database 	= "%[2]s"
			name 		= "dbo"
		}

		resource "azuresql_security_policy" "test" {
			database 	= "%[2]s"
			name        = "tfsecurity_policy_%[3]s"
			schema		= data.azuresql_schema.dbo.id
		}
		`, template, connection, name)
}

func (r SecurityPolicyResource) template() string {
	return fmt.Sprintf(`
		provider "azuresql" {
		}
	`)
}
