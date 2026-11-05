package sql

import (
	"context"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

type staticCredential struct{}

func (staticCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "token"}, nil
}

func TestWithoutFedAuth(t *testing.T) {
	tests := map[string]string{
		"sqlserver://server.database.windows.net:1433?database=db&fedauth=ActiveDirectoryDefault": "sqlserver://server.database.windows.net:1433?database=db",
		"sqlserver://server.database.windows.net:1433?fedauth=ActiveDirectoryDefault":             "sqlserver://server.database.windows.net:1433",
		"sqlserver://server.datawarehouse.fabric.microsoft.com?database=db":                       "sqlserver://server.datawarehouse.fabric.microsoft.com?database=db",
	}

	for input, expected := range tests {
		if actual := withoutFedAuth(input); actual != expected {
			t.Errorf("withoutFedAuth(%q) = %q, want %q", input, actual, expected)
		}
	}
}

func TestOpenWithCredential(t *testing.T) {
	cache := NewCache("", false, false)
	cache.Credential = staticCredential{}

	ctx := context.Background()
	for _, provider := range []string{"sqlserver::server:1433:db", "synapse::server:1433", "fabric::server:1433:db"} {
		connection := ParseConnectionId(ctx, provider)
		db, err := cache.open(connection.ConnectionString)
		if err != nil {
			t.Errorf("open(%s) returned error: %v", provider, err)
			continue
		}
		db.Close()
	}
}

func TestOpenWithoutCredential(t *testing.T) {
	cache := NewCache("", false, false)

	connection := ParseConnectionId(context.Background(), "sqlserver::server:1433:db")
	db, err := cache.open(connection.ConnectionString)
	if err != nil {
		t.Fatalf("open returned error: %v", err)
	}
	db.Close()
}
