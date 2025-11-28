package sql

import (
	"context"
	"fmt"
	"terraform-provider-azuresql/internal/logging"
)

func Execute(ctx context.Context, connection Connection, sql string) {

	_, err := connection.Connection.ExecContext(ctx, sql)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Query execution failed"), err)
	}
}
