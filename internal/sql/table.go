package sql

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"context"
	"terraform-provider-azuresql/internal/logging"
)

type Table struct {
	Id         string
	Connection string
	Name       string
	Schema     string
	SchemaName string
	ObjectId   int64
}

func tableFormatId(connectionId string, id int64) string {
	return fmt.Sprintf("%s/table/%d", connectionId, id)
}

func isTableId(id string) bool {
	return strings.Contains(id, "/table/")
}

func parseTableId(ctx context.Context, id string) (table Table) {
	s := strings.Split(id, "/table/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /table/ exactly once")
		return
	}

	table.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	object_id, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse table id")
		return
	}

	table.ObjectId = object_id
	table.Id = id

	return
}

func GetTableFromNameAndSchema(ctx context.Context, connection Connection, name string, schema string, requiresExist bool) (table Table) {

	var objectId, schemaId int64
	if schema == "" {
		schemaId = 1
	} else {
		schemaObj := ParseSchemaId(ctx, schema)
		if schemaObj.Connection != connection.ConnectionId {
			logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", schema, connection.ConnectionId))
			return
		}
		schemaId = schemaObj.SchemaId
	}

	if err := connection.Connection.QueryRowContext(ctx, "SELECT object_id FROM sys.tables where name = @name and schema_id = @schema_id",
		sql.Named("name", name), sql.Named("schema_id", schemaId)).Scan(&objectId); err != nil {
		logging.AddError(ctx, fmt.Sprintf("Reading of table with name %s from schema %s failed", name, schema), err)
		return
	}

	return Table{
		Id:         tableFormatId(connection.ConnectionId, objectId),
		Connection: connection.ConnectionId,
		Name:       name,
		Schema:     schema,
		ObjectId:   objectId,
	}

}

func GetTableFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (table Table) {

	table = parseTableId(ctx, id)

	if logging.HasError(ctx) {
		return
	}

	var name, schemaName string
	var schemaId int64

	query := "SELECT name, schema_id, schema_name(schema_id) FROM sys.tables where object_id = @id"
	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("id", table.ObjectId)).Scan(&name, &schemaId, &schemaName)

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Table not found", fmt.Sprintf("Table %s doesn't exist", id))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading table with %s failed", id), err)
		return
	}

	return Table{
		Id:         table.Id,
		Connection: connection.ConnectionId,
		Name:       name,
		Schema:     schemaFormatId(connection.ConnectionId, schemaId),
		SchemaName: schemaName,
		ObjectId:   table.ObjectId,
	}

}
