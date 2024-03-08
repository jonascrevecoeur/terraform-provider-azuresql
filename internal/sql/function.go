package sql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type FunctionArgument struct {
	Name string
	Type string
}

type FunctionProps struct {
	Arguments     []FunctionArgument
	ReturnType    string
	Executor      string
	Schemabinding bool
	Definition    string
}

type Function struct {
	Id         string
	Connection string
	Name       string
	ObjectId   int64
	Schema     string
	Raw        string
	Properties FunctionProps
}

func functionFormatId(connectionId string, objectId int64) string {
	return fmt.Sprintf("%s/function/%d", connectionId, objectId)
}

func ParseFunctionId(ctx context.Context, id string) (function Function) {
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

func CreateFunctionFromRaw(ctx context.Context, connection Connection, name string, schemaResourceId string, definition string) (function Function) {

	schema := GetSchemaFromId(ctx, connection, schemaResourceId, true)

	if logging.HasError(ctx) {
		return
	}

	regex := fmt.Sprintf("^[\\s]*create function %s.%s[\\s]*\\(", schema.Name, name)
	if match, _ := regexp.MatchString("(?i)"+regex, definition); !match {
		logging.AddError(ctx, "Function creation failed", fmt.Sprintf("Function defintion should contain 'create function %s.%s()'. The given definition was %s.", schema.Name, name, definition))
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

func buildFunctionQuery(name string, schemaName string, props FunctionProps) string {

	arguments := ""
	for index, argument := range props.Arguments {
		if index != 0 {
			arguments += ", "
		}
		arguments += "@" + argument.Name + " " + argument.Type
	}

	var execute_as string
	if strings.ToLower(props.Executor) == "caller" {
		execute_as = ""
	} else if slices.Contains([]string{"self", "owner"}, strings.ToLower(props.Executor)) {
		execute_as = fmt.Sprintf("with execute as %s", props.Executor)
	} else {
		execute_as = fmt.Sprintf("with execute as '%s'", props.Executor)
	}

	schemabinding := ""
	if props.Schemabinding {
		if execute_as != "" {
			schemabinding = fmt.Sprintf(", schemabinding")
		} else {
			schemabinding = fmt.Sprintf("with schemabinding")
		}
	}

	// if needed add begin/end block and return
	definition := props.Definition
	if strings.ToLower(props.ReturnType) != "table" {
		if match, _ := regexp.MatchString("^[\\s]begin", strings.ToLower(definition)); !match {
			if match, _ := regexp.MatchString("return", strings.ToLower(definition)); !match {
				definition = fmt.Sprintf(`BEGIN
return %s
END`, definition)
			} else {
				definition = fmt.Sprintf(`BEGIN
%s
END`, definition)
			}
		}
	} else {
		if match, _ := regexp.MatchString("return", strings.ToLower(definition)); !match {
			definition = fmt.Sprintf("return %s", definition)
		}
	}

	return fmt.Sprintf(`
create function %s.%s (%s)
returns %s
%s%s
as 
%s
`, schemaName, name, arguments, props.ReturnType, execute_as, schemabinding, definition)
}

func CreateFunctionFromProperties(ctx context.Context, connection Connection, name string, schemaResourceId string, props FunctionProps) (function Function) {

	schema := GetSchemaFromId(ctx, connection, schemaResourceId, true)

	if logging.HasError(ctx) {
		return
	}

	query := buildFunctionQuery(name, schema.Name, props)

	return CreateFunctionFromRaw(ctx, connection, name, schemaResourceId, query)
}

func GetFunctionFromNameAndSchema(ctx context.Context, connection Connection, name string, schemaResourceId string, requiresExist bool) (function Function) {
	schema := ParseSchemaId(ctx, schemaResourceId)

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

	function.Id = functionFormatId(connection.ConnectionId, objectId)
	function.ObjectId = objectId
	function.Schema = schemaFormatId(connection.ConnectionId, schemaId)
	function.Connection = connection.ConnectionId

	return
}

func GetFunctionFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (function Function) {
	function = ParseFunctionId(ctx, id)
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
