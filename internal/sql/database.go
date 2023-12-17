package sql

import (
	"context"
	"database/sql"
	"fmt"
	"terraform-provider-azuresql/internal/logging"
)

type Database struct {
	Id         string
	Connection string
	Name       string
}

func databaseFormatId(connectionId string, name string) string {
	return fmt.Sprintf("%s:%s", connectionId, name)
}

func CreateDatabase(ctx context.Context, connection Connection, name string) (database Database) {

	query := fmt.Sprintf("create database [%s]", name)

	_, err := connection.Connection.ExecContext(ctx, query)

	logging.AddError(ctx, fmt.Sprintf("Databse creation failed for database %s", name), err)

	database = GetDatabaseFromName(ctx, connection, name)
	if database.Id == "" && !logging.HasError(ctx) {
		logging.AddError(ctx, "Unable to read newly created database", fmt.Sprintf("Unable to read dabase %s after creation.", name))
	}

	return database
}

func GetDatabaseFromName(ctx context.Context, connection Connection, name string) (database Database) {

	var id int64

	query := "select database_id from sys.databases where name = @name"

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("name", name)).
		Scan(&id))

	switch {
	case err == sql.ErrNoRows:
		// database doesn't exist
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading database %s failed", name), err)
		return
	}

	return Database{
		Id:         databaseFormatId(connection.ConnectionId, name),
		Connection: connection.ConnectionId,
		Name:       name,
	}

}

func DropDatabase(ctx context.Context, connection Connection, name string) {

	// check if database exists
	database := GetDatabaseFromName(ctx, connection, name)
	if logging.HasError(ctx) || database.Id == "" {
		return
	}

	// drop all connections from the database before deletion
	var err error
	query := fmt.Sprintf(`
		DECLARE @kill varchar(8000) = '';  
		SELECT @kill = @kill + 'kill ' + CONVERT(varchar(5), session_id) + ';'  
		FROM sys.dm_exec_sessions
		WHERE database_id  = db_id('%s')
		
		exec(@kill)
	`, name)
	_, err = connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Closing connections from database %s failed", name), err)
		return
	}

	query = fmt.Sprintf("drop database if exists [%s]", name)
	_, err = connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping database %s failed", name), err)
	}
}
