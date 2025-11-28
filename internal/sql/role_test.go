package sql

import (
	"terraform-provider-azuresql/internal/logging"
	"testing"
)

func TestParseRoleId(t *testing.T) {
	ctx := logging.GetTestContext()

	tests := map[string]struct {
		expectedConnection  string
		expectedPrincipalId int64
		expectError         bool
	}{
		"sqlserver::server:1433/role/11": {
			expectedConnection:  "sqlserver::server:1433",
			expectedPrincipalId: 11,
			expectError:         false,
		},
	}

	for roleId, expected := range tests {
		actual := ParseRoleId(ctx, roleId)

		if logging.HasError(ctx) && !expected.expectError {
			t.Errorf("Parsing roleId %s should not throw an error", roleId)
			continue
		}
		if !logging.HasError(ctx) && expected.expectError {
			t.Errorf("Parsing roleId %s should throw an error", roleId)
			continue
		}

		if actual.Connection != expected.expectedConnection {
			t.Errorf("Expected connectionString %s does not match actual %s when parsing roleId %s",
				expected.expectedConnection, actual.Connection, roleId)
		}

		if actual.PrincipalId != expected.expectedPrincipalId {
			t.Errorf("Expected principal id %d does not match actual %d when parsing roleId %s",
				expected.expectedPrincipalId, actual.PrincipalId, roleId)
		}

		logging.ClearDiagnostics(ctx)
	}
}
