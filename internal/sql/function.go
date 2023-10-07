package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type Function struct {
	Id         string
	Connection string
	Name       string
	ObjectId   int64
	Schema     string
	Raw        string
}

func functionFormatId(connectionId string, objectId int64) string {
	return fmt.Sprintf("%s/function/%d", connectionId, objectId)
}

func parseFunctionId(ctx context.Context, id string) (function Function) {
	s := strings.Split(id, "/function/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /function/ exactly once")
		return
	}

	function.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	objectId, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse function id")
		return
	}

	function.ObjectId = objectId

	return
}

func CreateFunctionFromDefinition(ctx context.Context, connection Connection, name string, schemaResourceId string, definition string) (function Function) {

	schema := GetSchemaFromId(ctx, connection, schemaResourceId, true)

	if logging.HasError(ctx) {
		return
	}

	query_start := strings.ToLower(fmt.Sprintf("create function %s.%s(", schema.Name, name))
	if !strings.Contains(strings.ToLower(definition), query_start) {
		logging.AddError(ctx, "Function creation failed", fmt.Sprintf("Function defintion should contain '%s'", query_start))
		return
	}

	_, err := connection.Connection.ExecContext(ctx, definition)

	if err != nil {
		logging.AddError(ctx, "Function creation failed", err)
		return
	}

	function = GetFunctionFromNameAndSchema(ctx, connection, name, schemaResourceId, false)
	if !logging.HasError(ctx) && function.Id == "" {
		logging.AddError(ctx, "Unable to read newly created function", fmt.Sprintf("Unable to read function %s after creation.", name))
	}

	return function
}

func GetFunctionFromNameAndSchema(ctx context.Context, connection Connection, name string, schemaResourceId string, requiresExist bool) (function Function) {
	schema := parseSchemaId(ctx, schemaResourceId)

	query := `
		select obj.object_id, mod.definition from sys.objects obj 
		inner join sys.sql_modules mod
		on obj.object_id = mod.object_id
		where obj.name = @name and obj.schema_id = @schema_id 
		and type in ('IF', 'FN', 'TF')`

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("name", name), sql.Named("schema_id", schema.SchemaId)).Scan(&function.ObjectId, &function.Raw)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Function not found", fmt.Sprintf("Function with name %s doesn't exist", name))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading function %s failed", name), err)
		return
	}

	function.Id = functionFormatId(connection.ConnectionId, function.ObjectId)
	function.Schema = schemaResourceId
	function.Name = name

	return
}

func GetFunctionFromObjectId(ctx context.Context, connection Connection, objectId int64, requiresExist bool) (function Function) {
	var schemaId int64
	query := `
		select obj.schema_id, obj.name, mod.definition from sys.objects obj 
		inner join sys.sql_modules mod
		on obj.object_id = mod.object_id
		where obj.object_id = @object_id
		and type in ('IF', 'FN', 'TF')`

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("object_id", objectId)).Scan(&schemaId, &function.Name, &function.Raw)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Function not found", fmt.Sprintf("Function with objectId %d doesn't exist", objectId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading function %d failed", objectId), err)
		return
	}

	function.Id = functionFormatId(connection.ConnectionId, function.ObjectId)
	function.ObjectId = objectId
	function.Schema = schemaFormatId(connection.ConnectionId, schemaId)

	return
}

func GetFunctionFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (function Function) {
	function = parseFunctionId(ctx, id)
	if logging.HasError(ctx) {
		return
	}

	if function.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	function = GetFunctionFromObjectId(ctx, connection, function.ObjectId, requiresExist)

	return function
}

func DropFunction(ctx context.Context, connection Connection, id string) {

	function := GetFunctionFromId(ctx, connection, id, false)
	if logging.HasError(ctx) || function.Id == "" {
		return
	}
	schema := GetSchemaFromId(ctx, connection, function.Schema, true)
	if logging.HasError(ctx) {
		return
	}

	query := fmt.Sprintf("drop function %s.%s", schema.Name, function.Name)

	if _, err := connection.Connection.ExecContext(ctx, query); err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping function %s.%s failed", schema.Name, function.Name), err)
	}
}
