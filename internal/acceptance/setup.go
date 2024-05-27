package acceptance

import (
	"terraform-provider-azuresql/internal/provider"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const ProviderConfig = `
provider "azuresql" {
	
}
`

var (
	TestAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"azuresql": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
)
