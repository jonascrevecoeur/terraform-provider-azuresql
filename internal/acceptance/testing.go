package acceptance

import (
	"log"
	"strings"
	"testing"

	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"
)

var commonEnv = []string{
	"AZURE_AD_USER",
	"AZURE_AD_GROUP",
	"AZURE_SUBSCRIPTION",
	"AZURE_CLIENT_ID_OPT",
	"AZURE_CLIENT_SECRET_OPT",
}

func PreCheck(t *testing.T) {
	for _, backend := range Backends(t) {
		PreCheckBackend(t, backend)
	}
}

func PreCheckBackend(t *testing.T, backend Backend) {
	requireEnv(t, commonEnv...)
	requireEnv(t, backend.requiredEnv...)
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
