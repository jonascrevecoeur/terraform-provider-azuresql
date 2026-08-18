package sql

import (
	"testing"

	"terraform-provider-azuresql/internal/logging"

	"github.com/stretchr/testify/assert"
)

func TestParseUserDefaultSchemaId(t *testing.T) {
	ctx := logging.GetTestContext()
	setting := ParseUserDefaultSchemaId(ctx, "sqlserver::server:1433:database/user_default_schema/42")

	assert.Equal(t, "sqlserver::server:1433:database/user_default_schema/42", setting.Id)
	assert.Equal(t, "sqlserver::server:1433:database", setting.Connection)
	assert.Equal(t, "sqlserver::server:1433:database/user/42", setting.User)
}

func TestQuoteIdentifier(t *testing.T) {
	assert.Equal(t, "[dbo]", quoteIdentifier("dbo"))
	assert.Equal(t, "[group]]name]", quoteIdentifier("group]name"))
}
