package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type DatabaseScopedCredential struct {
	Id           string
	Connection   string
	Name         string
	Identity     string
	Secret       string
	CredentialId int64
}

func databaseScopedCredentialFormatId(connectionId string, credentialId int64) string {
	return fmt.Sprintf("%s/databasescopedcredential/%d", connectionId, credentialId)
}

// retrieve name and sid from a tf databaseScopedCredential id
func ParseDatabaseScopedCredentialId(ctx context.Context, id string) (databaseScopedCredential DatabaseScopedCredential) {
	s := strings.Split(id, "/databasescopedcredential/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /databasescopedcredential/ exactly once")
		return
	}

	databaseScopedCredential.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	credentialId, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse database scoped credential id")
		return
	}

	databaseScopedCredential.CredentialId = credentialId
	return
}

func CreateDatabaseScopedCredential(ctx context.Context, connection Connection, name string, identity string, secret string) (databaseScopedCredential DatabaseScopedCredential) {

	var query string
	if secret == "" {
		query = fmt.Sprintf("create database scoped credential %s with identity = '%s'", name, identity)
	} else {
		query = fmt.Sprintf("create database scoped credential %s with identity = '%s', secret='%s'", name, identity, secret)
	}

	_, err := connection.Connection.ExecContext(ctx, query)
	logging.AddError(ctx, "Creation of database scoped credential failed", err)

	// set requiresExist to false in order to specify a custom error message
	databaseScopedCredential = GetDatabaseScopedCredentialFromName(ctx, connection, name, false)
	if !logging.HasError(ctx) && databaseScopedCredential.Id == "" {
		logging.AddError(ctx, "Unable to read newly created database scoped credential", fmt.Sprintf("Unable to read database scoped credential %s after creation.", name))
	}

	databaseScopedCredential.Secret = secret
	return
}

func AlterDatabaseScopedCredential(ctx context.Context, connection Connection, name string, identity string, secret string) (databaseScopedCredential DatabaseScopedCredential) {

	var query string
	if secret == "" {
		query = fmt.Sprintf("alter database scoped credential %s with identity = '%s'", name, identity)
	} else {
		query = fmt.Sprintf("alter database scoped credential %s with identity = '%s', secret='%s'", name, identity, secret)
	}

	_, err := connection.Connection.ExecContext(ctx, query)
	logging.AddError(ctx, "Updating database scoped credential failed", err)

	// set requiresExist to false in order to specify a custom error message
	databaseScopedCredential = GetDatabaseScopedCredentialFromName(ctx, connection, name, false)
	if !logging.HasError(ctx) && databaseScopedCredential.Id == "" {
		logging.AddError(ctx, "Unable to read updated database scoped credential", fmt.Sprintf("Unable to read database scoped credential %s after update.", name))
	}

	databaseScopedCredential.Secret = secret
	return
}

func GetDatabaseScopedCredentialFromName(ctx context.Context, connection Connection, name string, requiresExist bool) (databaseScopedCredential DatabaseScopedCredential) {
	var credentialId int64
	var identity string

	query := `
		select credential_id, credential_identity from sys.database_scoped_credentials
		where name=@name`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("name", name)).
		Scan(&credentialId, &identity))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Database scoped credential not found", fmt.Sprintf("Database scoped credential with name %s doesn't exist", name))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading database scoped credential %s failed", name), err)
		return
	}

	return DatabaseScopedCredential{
		Id:           databaseScopedCredentialFormatId(connection.ConnectionId, credentialId),
		Connection:   connection.ConnectionId,
		Name:         name,
		Identity:     identity,
		CredentialId: credentialId,
	}
}

func GetDatabaseScopedCredentialFromCredentialId(ctx context.Context, connection Connection, credentialId int64, requiresExist bool) (databaseScopedCredential DatabaseScopedCredential) {
	var name, identity string

	query := `
		select name, credential_identity from sys.database_scoped_credentials
		where credential_id=@credential`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("credential", credentialId)).
		Scan(&name, &identity))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Database scoped credential not found", fmt.Sprintf("Database scoped credential with credential id %d doesn't exist", credentialId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading database scoped credential with credential id %d failed", credentialId), err)
		return
	}

	return DatabaseScopedCredential{
		Id:           databaseScopedCredentialFormatId(connection.ConnectionId, credentialId),
		Connection:   connection.ConnectionId,
		Name:         name,
		Identity:     identity,
		CredentialId: credentialId,
	}
}

func GetDatabaseScopedCredentialFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (databaseScopedCredential DatabaseScopedCredential) {
	databaseScopedCredential = ParseDatabaseScopedCredentialId(ctx, id)
	if logging.HasError(ctx) {
		return
	}

	if databaseScopedCredential.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	return GetDatabaseScopedCredentialFromCredentialId(ctx, connection, databaseScopedCredential.CredentialId, requiresExist)
}

func DropDatabaseScopedCredential(ctx context.Context, connection Connection, id string) {

	databaseScopedCredential := GetDatabaseScopedCredentialFromId(ctx, connection, id, false)

	if logging.HasError(ctx) || databaseScopedCredential.Id == "" {
		return
	}

	var err error
	_, err = connection.Connection.ExecContext(ctx, fmt.Sprintf("drop database scoped credential %s", databaseScopedCredential.Name))

	if err != nil {
		logging.AddError(ctx, "Dropping database scoped credential failed", err)
	}
}
