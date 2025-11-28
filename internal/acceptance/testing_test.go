package acceptance

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"terraform-provider-azuresql/internal/logging"
	"terraform-provider-azuresql/internal/sql"
	"testing"

	_ "github.com/microsoft/go-mssqldb/azuread"
)

func TestAzureCLI(t *testing.T) {
	commandLine := "az --version"
	cliCmd := exec.CommandContext(logging.GetTestContext(), "/bin/sh", "-c", commandLine)

	var stderr bytes.Buffer
	cliCmd.Stderr = &stderr

	_, err := cliCmd.Output()
	if err != nil {
		msg := stderr.String()
		var exErr *exec.ExitError
		t.Error(msg)
		if errors.As(err, &exErr) && exErr.ExitCode() == 127 || strings.HasPrefix(msg, "'az' is not recognized") {
			t.Error("Azure CLI not found on path")

		}
		if msg == "" {
			t.Error(err.Error())
		}
	}
}

func TestExecuteSQL(t *testing.T) {
	PreCheck(t)
	data := BuildTestData(t)

	ExecuteSQL(data.SQLDatabase_connection, fmt.Sprintf("create table dbo.tftable_%s (col1 int)", data.RandomString))
	defer ExecuteSQL(data.SQLDatabase_connection, fmt.Sprintf("DROP table dbo.tftable_%s", data.RandomString))
}

func TestSQLServerExists(t *testing.T) {
	if os.Getenv("AZURE_SUBSCRIPTION") == "" {
		log.Fatal("Environment variable AZURE_SUBSCRIPTION has to be set for this test")
	}

	ctx := logging.GetTestContext()
	cache := sql.NewCache(os.Getenv("AZURE_SUBSCRIPTION"), true, false)
	data := BuildTestData(t)

	connection := sql.ParseConnectionId(ctx, data.SQLServer_connection)

	status := cache.ServerExists(ctx, connection)

	if logging.HasError(ctx) {
		for _, err := range logging.GetDiagnostics(ctx).Errors() {
			log.Fatalf("%s - %s", err.Summary(), err.Detail())
		}
	}

	if status != sql.ConnectionResourceStatusExists {
		log.Fatal(fmt.Sprintf("Wrong status (%d) returned by ServerExists, expected ConnectionResourceStatusExists", status))
	}

	connection.Server = "servernotexist"

	status = cache.ServerExists(ctx, connection)

	if logging.HasError(ctx) {
		for _, err := range logging.GetDiagnostics(ctx).Errors() {
			log.Fatalf("%s - %s", err.Summary(), err.Detail())
		}
	}

	if status != sql.ConnectionResourceStatusNotFound {
		log.Fatal("Wrong status returned by ServerExists, expected ConnectionResourceStatusNotFound")
	}
}

func TestSQLDatabaseExists(t *testing.T) {
	if os.Getenv("AZURE_SUBSCRIPTION") == "" {
		log.Fatal("Environment variable AZURE_SUBSCRIPTION has to be set for this test")
	}

	ctx := logging.GetTestContext()
	cache := sql.NewCache(os.Getenv("AZURE_SUBSCRIPTION"), true, true)
	data := BuildTestData(t)

	connection := sql.ParseConnectionId(ctx, data.SQLDatabase_connection)

	status := cache.DatabaseExists(ctx, connection)

	if logging.HasError(ctx) {
		for _, err := range logging.GetDiagnostics(ctx).Errors() {
			log.Fatalf("%s - %s", err.Summary(), err.Detail())
		}
	}

	if status != sql.ConnectionResourceStatusExists {
		log.Fatal(fmt.Sprintf("Wrong status (%d) returned by DatabaseExists, expected ConnectionResourceStatusExists", status))
	}

	connection.Database += "notexist"
	connection.ConnectionId += "notexist"

	status = cache.DatabaseExists(ctx, connection)

	if logging.HasError(ctx) {
		for _, err := range logging.GetDiagnostics(ctx).Errors() {
			log.Fatalf("%s - %s", err.Summary(), err.Detail())
		}
	}

	if status != sql.ConnectionResourceStatusNotFound {
		log.Fatal("Wrong status returned by DatabaseExists, expected ConnectionResourceStatusNotFound")
	}

	connection = sql.ParseConnectionId(ctx, data.SQLDatabase_connection)
	connection.ConnectionId = strings.Replace(connection.ConnectionId, connection.Server, "servernotexist", 1)
	connection.Server = "servernotexist"

	status = cache.DatabaseExists(ctx, connection)

	if logging.HasError(ctx) {
		for _, err := range logging.GetDiagnostics(ctx).Errors() {
			log.Fatalf("%s - %s", err.Summary(), err.Detail())
		}
	}

	if status != sql.ConnectionResourceStatusNotFound {
		log.Fatal("Wrong status returned by DatabaseExists, expected ConnectionResourceStatusNotFound")
	}
}
