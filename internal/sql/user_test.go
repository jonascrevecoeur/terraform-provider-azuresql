package sql

import (
	"terraform-provider-azuresql/internal/logging"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserTypeCode(t *testing.T) {
	ctx := logging.GetTestContext()
	assert.Equal(t, userTypeCode(ctx, ""), "E")
	assert.Equal(t, userTypeCode(ctx, "AD user"), "E")
	assert.Equal(t, userTypeCode(ctx, "AD group"), "X")
	assert.False(t, logging.HasError(ctx))

	assert.Equal(t, userTypeCode(ctx, "SQL user"), "")
	assert.True(t, logging.HasError(ctx))
}
