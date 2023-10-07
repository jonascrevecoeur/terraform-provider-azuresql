package acceptance

import (
	"log"
	"os"
	"strings"
	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"
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

func ExecuteSQL(connectionId string, query string) {
	cache := sql.NewCache()

	isServer := len(strings.Split(connectionId, ":")) == 5
	connection := cache.Connect(logging.GetTestContext(), connectionId, isServer)

	_, err := connection.Connection.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
