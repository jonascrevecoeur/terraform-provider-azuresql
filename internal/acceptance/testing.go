package acceptance

import (
	"os"
	"testing"
)

func PreCheck(t *testing.T) {
	variables := []string{
		"AZURE_SQL_SERVER",
		"AZURE_SQL_SERVER_PORT",
		"AZURE_SQL_DATABASE",
		"AZURE_SYNAPSE_SERVER",
		"AZURE_SYNAPSE_DATABASE",
		"AZURE_SYNAPSE_SERVER_PORT",
		"AZURE_AD_USER",
		"AZURE_AD_GROUP",
	}

	for _, variable := range variables {
		value := os.Getenv(variable)
		if value == "" {
			t.Fatalf("`%s` must be set for acceptance tests!", variable)
		}
	}
}
