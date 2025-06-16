package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type MasterKey struct {
	Id         string
	Connection string
	Password   string
}

func masterKeyFormatId(connectionId string) string {
	return fmt.Sprintf("%s/masterkey", connectionId)
}

// retrieve name and sid from a tf masterKey id
func ParseMasterKeyId(ctx context.Context, id string) (masterKey MasterKey) {
	if !strings.HasSuffix(id, "/masterkey") {
		logging.AddError(ctx, "ID format error", "id doesn't ends in /masterKey")
		return
	}

	masterKey.Id = id
	masterKey.Connection = strings.TrimSuffix(id, "/masterkey")

	return
}

func CreateMasterKey(ctx context.Context, connection Connection) (masterKey MasterKey) {
	password := GeneratePassword(20, 3, 4, 5)

	query := fmt.Sprintf("create master key encryption by password = '%s'", password)

	_, err := connection.Connection.ExecContext(ctx, query)
	logging.AddError(ctx, "Creation of master key failed", err)

	return MasterKey{
		Id:         masterKeyFormatId(connection.ConnectionId),
		Connection: connection.ConnectionId,
		Password:   password,
	}
}

func MasterKeyExists(ctx context.Context, connection Connection) bool {

	query := fmt.Sprintf("select 1 from sys.symmetric_keys where name = '##MS_DatabaseMasterKey##'")

	var x int
	err := (connection.
		Connection.
		QueryRowContext(ctx, query).
		Scan(&x))

	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		logging.AddError(ctx, "Checking if master key exists failed", err)
		return false
	}

	return true
}

func DropMasterKey(ctx context.Context, connection Connection) {

	hasMasterKey := MasterKeyExists(ctx, connection)

	if logging.HasError(ctx) || !hasMasterKey {
		return
	}

	var err error
	_, err = connection.Connection.ExecContext(ctx, "drop master key")

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping master key failed for database %s", connection.ConnectionId), err)
	}
}
