package sql

import (
	"fmt"
	"os"
	"terraform-provider-azuresql/internal/logging"
	"testing"

	_ "github.com/microsoft/go-mssqldb/azuread"
)

func TestParseConnectionId(t *testing.T) {
	ctx := logging.GetTestContext()

	tests := map[string]struct {
		expectecOutcome Connection
		expectError     bool
	}{
		"sqlserver::server:150:db": {
			expectecOutcome: Connection{
				IsServerConnection: false,
				Provider:           "sqlserver",
				ConnectionString:   "sqlserver://server.database.windows.net:150?database=db&fedauth=ActiveDirectoryDefault",
			},
			expectError: false,
		},
		"sqlserver::server:150": {
			expectecOutcome: Connection{
				IsServerConnection: true,
				Provider:           "sqlserver",
				ConnectionString:   "sqlserver://server.database.windows.net:150?fedauth=ActiveDirectoryDefault",
			},
			expectError: false,
		},
		"sqlserver::server": {
			expectecOutcome: Connection{},
			expectError:     true,
		},
	}

	for connectionId, expected := range tests {
		actual := ParseConnectionId(ctx, connectionId)

		if logging.HasError(ctx) && !expected.expectError {
			t.Errorf("Parsing connectionId %s should not throw an error", connectionId)
			continue
		}
		if !logging.HasError(ctx) && expected.expectError {
			t.Errorf("Parsing connectionId %s should throw an error", connectionId)
			continue
		}

		if actual.ConnectionString != expected.expectecOutcome.ConnectionString {
			t.Errorf("Expected connectionString %s does not match actual %s when parsing connectionId %s",
				expected.expectecOutcome.ConnectionString, actual.ConnectionString, connectionId)
		}

		if actual.Provider != expected.expectecOutcome.Provider {
			t.Errorf("Expected provider %s does not match actual %s when parsing connectionId %s",
				expected.expectecOutcome.Provider, actual.Provider, connectionId)
		}

		if actual.IsServerConnection != expected.expectecOutcome.IsServerConnection {
			t.Errorf("Expected type %t does not match actual %t when parsing connectionId %s",
				expected.expectecOutcome.IsServerConnection, actual.IsServerConnection, connectionId)
		}

		logging.ClearDiagnostics(ctx)
	}
}

func TestRecoverClosedConnection(t *testing.T) {
	ctx := logging.GetTestContext()
	connection_id := fmt.Sprintf("sqlserver::%s:%s", os.Getenv("AZURE_SQL_SERVER"), os.Getenv("AZURE_SQL_SERVER_PORT"))

	fmt.Printf("t: %v\n", connection_id)

	cache := NewCache(
		os.Getenv("AZURE_SUBSCRIPTION"),
		false,
		false,
	)

	connection := cache.Connect(ctx, connection_id, true, true)
	var result int64
	connection.Connection.Close()

	connection = cache.Connect(ctx, connection_id, true, true)
	err := connection.Connection.QueryRow("select 1 as a").Scan(&result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

}
