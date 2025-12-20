package acceptance

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"
	"testing"

	"github.com/joho/godotenv"
)

func PreCheck(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Could not get current file path")
	}
	err := godotenv.Overload(filepath.Join(filepath.Dir(filename), "../../.env"))
	if err != nil {
		t.Fatal("No .env file found")
	}

	variables := []string{
		"AZURE_SQL_SERVER",
		"AZURE_SQL_SERVER_PORT",
		"AZURE_SQL_DATABASE",
		"AZURE_SYNAPSE_SERVER",
		"AZURE_SYNAPSE_DATABASE",
		"AZURE_SYNAPSE_SERVER_PORT",
		"AZURE_FABRIC_SERVER",
		"AZURE_FABRIC_DATABASE",
		"AZURE_AD_USER",
		"AZURE_AD_GROUP",
		"AZURE_SUBSCRIPTION",
		"AZURE_CLIENT_ID_OPT",
		"AZURE_CLIENT_SECRET_OPT",
	}

	for _, variable := range variables {
		value := os.Getenv(variable)
		if value == "" {
			t.Fatalf("`%s` must be set for acceptance tests!", variable)
		}
	}
}

func ExecuteSQL(connectionId string, query string) {
	cache := sql.NewCache("", false, false)

	isServer := len(strings.Split(connectionId, ":")) == 5
	connection := cache.Connect(logging.GetTestContext(), connectionId, isServer, true)

	_, err := connection.Connection.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}
