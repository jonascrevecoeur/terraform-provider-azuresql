package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"terraform-provider-azuresql/internal/logging"
)

// UserDefaultSchema is the DEFAULT_SCHEMA setting for a database user.
type UserDefaultSchema struct {
	Id            string
	Connection    string
	User          string
	UserName      string
	DefaultSchema string
}

func userDefaultSchemaFormatId(connectionId string, principalId int64) string {
	return fmt.Sprintf("%s/user_default_schema/%d", connectionId, principalId)
}

func ParseUserDefaultSchemaId(ctx context.Context, id string) (setting UserDefaultSchema) {
	parts := strings.Split(id, "/user_default_schema/")
	if len(parts) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /user_default_schema/ exactly once")
		return
	}

	principalId, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", fmt.Sprintf("Unable to parse user default schema id %s", id))
		return
	}

	setting.Connection = parts[0]
	setting.Id = id
	setting.User = userFormatId(setting.Connection, principalId)
	return setting
}

// quoteIdentifier quotes a SQL Server identifier. Identifiers cannot be
// supplied as query parameters, so closing brackets must be escaped before
// interpolating them into DDL.
func quoteIdentifier(identifier string) string {
	return "[" + strings.ReplaceAll(identifier, "]", "]]") + "]"
}

func SetUserDefaultSchema(ctx context.Context, connection Connection, userName string, defaultSchema string) {
	if strings.TrimSpace(userName) == "" || strings.TrimSpace(defaultSchema) == "" {
		logging.AddError(ctx, "Invalid user default schema", "user name and default schema must not be empty")
		return
	}

	query := fmt.Sprintf("ALTER USER %s WITH DEFAULT_SCHEMA = %s", quoteIdentifier(userName), quoteIdentifier(defaultSchema))
	if _, err := connection.Connection.ExecContext(ctx, query); err != nil {
		logging.AddError(ctx, fmt.Sprintf("Setting default schema for user %s failed", userName), err)
	}
}

func GetUserDefaultSchema(ctx context.Context, connection Connection, userResourceId string, requiresExist bool) (setting UserDefaultSchema) {
	user := GetUserFromId(ctx, connection, userResourceId, requiresExist)
	if logging.HasError(ctx) || user.Id == "" {
		return
	}

	var defaultSchema sql.NullString
	query := `
		select default_schema_name
		from sys.database_principals
		where principal_id = @principal_id and type != 'R'`

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("principal_id", user.PrincipalId)).Scan(&defaultSchema)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "User not found", fmt.Sprintf("User with id %s doesn't exist", userResourceId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading default schema for user %s failed", user.Name), err)
		return
	}

	return UserDefaultSchema{
		Id:            userDefaultSchemaFormatId(connection.ConnectionId, user.PrincipalId),
		Connection:    connection.ConnectionId,
		User:          user.Id,
		UserName:      user.Name,
		DefaultSchema: defaultSchema.String,
	}
}
