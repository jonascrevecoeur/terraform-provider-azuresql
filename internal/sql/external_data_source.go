package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type ExternalDataSource struct {
	Id           string
	Connection   string
	Name         string
	DataSourceId int64
	Credential   string
	Location     string
}

func externalDataSourceFormatId(connectionId string, dataSourceId int64) string {
	return fmt.Sprintf("%s/externaldatasource/%d", connectionId, dataSourceId)
}

func externalDataSourceFormatCredential(connectionId string, credentialId int64) string {
	if credentialId == 0 {
		return ""
	} else {
		return databaseScopedCredentialFormatId(connectionId, credentialId)
	}
}

func isExternalDataSourceId(id string) (isExternalDataSource bool) {
	return strings.Contains(id, "/externaldatasource/")
}

func ParseExternalDataSourceId(ctx context.Context, id string) (externalDataSource ExternalDataSource) {
	s := strings.Split(id, "/externaldatasource/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /externaldatasource/ exactly once")
		return
	}

	externalDataSource.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	dataSourceId, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", fmt.Sprintf("Unable to parse external data source id %s", id))
		return
	}

	externalDataSource.DataSourceId = dataSourceId

	return
}

func CreateExternalDataSource(ctx context.Context, connection Connection, name string, location string, credential string) (externalDataSource ExternalDataSource) {

	var credential_arg, type_arg string
	if credential != "" {
		databaseScopedCredential := GetDatabaseScopedCredentialFromId(ctx, connection, credential, true)

		if logging.HasError(ctx) {
			return
		}

		credential_arg = fmt.Sprintf(", credential = %s", databaseScopedCredential.Name)
	} else {
		credential_arg = ""
	}

	if connection.Provider == "synapse" {
		type_arg = ""
	} else {
		type_arg = ", type = BLOB_STORAGE"
	}

	query := fmt.Sprintf(`
		create external data source [%s]
		with (location = '%s' %s %s)
		`, name, location, credential_arg, type_arg)

	_, err := connection.Connection.ExecContext(ctx, query)

	logging.AddError(ctx, fmt.Sprintf("External data source creation failed for %s", name), err)

	// set requiresExist to false in order to specify a custom error message
	externalDataSource = GetExternalDataSourceFromName(ctx, connection, name, false)
	if !logging.HasError(ctx) && externalDataSource.Id == "" {
		logging.AddError(ctx, "Unable to read newly created external data source", fmt.Sprintf("Unable to read external data source %s after creation.", name))
	}

	return externalDataSource
}

func GetExternalDataSourceFromName(ctx context.Context, connection Connection, name string, requiresExist bool) (externalDataSource ExternalDataSource) {

	var dataSourceId, credentialId int64
	var location string

	query := `
		select data_source_id, credential_id, location
		from sys.external_data_sources
		where name = @name`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("name", name)).
		Scan(&dataSourceId, &credentialId, &location))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "External data source not found", fmt.Sprintf("External data source with name %s doesn't exist", name))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading external data source %s failed", name), err)
		return
	}

	return ExternalDataSource{
		Id:           externalDataSourceFormatId(connection.ConnectionId, dataSourceId),
		Connection:   connection.ConnectionId,
		Name:         name,
		Location:     location,
		Credential:   externalDataSourceFormatCredential(connection.ConnectionId, credentialId),
		DataSourceId: dataSourceId,
	}

}

func GetExternalDataSourceFromDataSourceId(ctx context.Context, connection Connection, dataSourceId int64, requiresExist bool) (externalDataSource ExternalDataSource) {

	var credentialId int64
	var location, name string

	query := `
		select name, credential_id, location
		from sys.external_data_sources
		where data_source_id = @id`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("id", dataSourceId)).
		Scan(&name, &credentialId, &location))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "External data source not found", fmt.Sprintf("External data source with data source id %d doesn't exist", dataSourceId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading external data source %s failed", name), err)
		return
	}

	return ExternalDataSource{
		Id:           externalDataSourceFormatId(connection.ConnectionId, dataSourceId),
		Connection:   connection.ConnectionId,
		Name:         name,
		Location:     location,
		Credential:   externalDataSourceFormatCredential(connection.ConnectionId, credentialId),
		DataSourceId: dataSourceId,
	}
}

// Get user from the azuresql terraform id
// requiresExist: Raise an error when the user doesn't exist
func GetExternalDataSourceFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (externalDataSource ExternalDataSource) {
	externalDataSource = ParseExternalDataSourceId(ctx, id)
	if logging.HasError(ctx) {
		return
	}

	if externalDataSource.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	externalDataSource = GetExternalDataSourceFromDataSourceId(ctx, connection, externalDataSource.DataSourceId, requiresExist)

	return externalDataSource
}

func DropExternalDataSource(ctx context.Context, connection Connection, dataSourceId int64) {

	externalDataSource := GetExternalDataSourceFromDataSourceId(ctx, connection, dataSourceId, false)
	if logging.HasError(ctx) || externalDataSource.Id == "" {
		return
	}

	query := fmt.Sprintf("drop external data source [%s]", externalDataSource.Name)
	var err error
	_, err = connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping exterinal data source %s failed", externalDataSource.Name), err)
	}
}
