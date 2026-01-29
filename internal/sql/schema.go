package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type Schema struct {
	Id         string
	Connection string
	Name       string
	SchemaId   int64
	Owner      string
}

func schemaFormatId(connectionId string, schemaId int64) string {
	return fmt.Sprintf("%s/schema/%d", connectionId, schemaId)
}

func isSchemaId(id string) bool {
	return strings.Contains(id, "/schema/")
}

func ParseSchemaId(ctx context.Context, id string) (schema Schema) {
	s := strings.Split(id, "/schema/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /schema/ exactly once")
		return
	}

	schema.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	schemaId, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse schema id")
		return
	}

	schema.SchemaId = schemaId

	return
}

func CreateSchema(ctx context.Context, connection Connection, name string, owner string) (schema Schema) {

	query := fmt.Sprintf("create schema [%s]", name)

	if owner != "" {
		ownerPrincipal := GetPrincipalFromId(ctx, connection, owner, true)

		if logging.HasError(ctx) {
			return
		}

		query += fmt.Sprintf(" authorization [%s]", ownerPrincipal.Name)
	}

	_, err := connection.Connection.ExecContext(ctx, query)
	logging.AddError(ctx, fmt.Sprintf("Schema creation failed for schema %s", name), err)

	// set requiresExist to false in order to specify a custom error message
	schema = GetSchemaFromName(ctx, connection, name, false)
	if !logging.HasError(ctx) && schema.Id == "" {
		logging.AddError(ctx, "Unable to read newly created schema", fmt.Sprintf("Unable to read schema %s after creation.", name))
	}

	return schema
}

func GetSchemaFromName(ctx context.Context, connection Connection, name string, requiresExist bool) (schema Schema) {

	var id, ownerId int64
	var ownerType string

	query := `
		select schemas.schema_id, schemas.principal_id, owner.type as owner_type from sys.database_principals owner
		left join sys.schemas schemas on owner.principal_id = schemas.principal_id
		where schemas.name = @name`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("name", name)).
		Scan(&id, &ownerId, &ownerType))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Schema not found", fmt.Sprintf("Schema with name %s doesn't exist", name))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading schema %s failed", name), err)
		return
	}

	return Schema{
		Id:         schemaFormatId(connection.ConnectionId, id),
		Connection: connection.ConnectionId,
		Name:       name,
		SchemaId:   id,
		Owner:      principalFormatId(connection.ConnectionId, ownerId, ownerType),
	}

}

func GetSchemaFromSchemaId(ctx context.Context, connection Connection, schemaId int64, requiresExist bool) (schema Schema) {

	var ownerId int64
	var name, ownerType string

	query := `
		select schemas.name, schemas.principal_id, owner.type as owner_type from sys.database_principals owner
		left join sys.schemas schemas on owner.principal_id = schemas.principal_id
		where schemas.schema_id = @id`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("id", schemaId)).
		Scan(&name, &ownerId, &ownerType))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Schema not found", fmt.Sprintf("Schema with id %d doesn't exist", schemaId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading schema with id %d failed", schemaId), err)
		return
	}

	return Schema{
		Id:         schemaFormatId(connection.ConnectionId, schemaId),
		Connection: connection.ConnectionId,
		Name:       name,
		SchemaId:   schemaId,
		Owner:      principalFormatId(connection.ConnectionId, ownerId, ownerType),
	}

}

// Get user from the azuresql terraform id
// requiresExist: Raise an error when the user doesn't exist
func GetSchemaFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (schema Schema) {
	schema = ParseSchemaId(ctx, id)
	if logging.HasError(ctx) {
		return
	}

	if schema.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	return GetSchemaFromSchemaId(ctx, connection, schema.SchemaId, requiresExist)
}

func UpdateSchemaOwner(ctx context.Context, connection Connection, id string, owner string) {
	schema := GetSchemaFromId(ctx, connection, id, true)
	if logging.HasError(ctx) {
		return
	}

	ownerPrincipal := GetPrincipalFromId(ctx, connection, owner, true)

	if logging.HasError(ctx) {
		return
	}

	query := fmt.Sprintf("alter authorization on schema::%s to %s", schema.Name, ownerPrincipal.Name)
	_, err := connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Alter owner for schema %s failed", schema.Name), err)
	}
}

func DropSchema(ctx context.Context, connection Connection, schemaId int64) {

	schema := GetSchemaFromSchemaId(ctx, connection, schemaId, false)
	if logging.HasError(ctx) || schema.Id == "" {
		return
	}

	query := fmt.Sprintf(`
		IF SCHEMA_ID('%[1]s') IS NOT NULL
		BEGIN
			DROP SCHEMA [%[1]s]
		END
	`, schema.Name)

	var err error
	_, err = connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping schema %s failed", schema.Name), err)
	}
}
