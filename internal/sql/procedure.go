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

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ProcedureArgument struct {
	Name string
	Type string
}

type ProcedureProps struct {
	Arguments     []ProcedureArgument
	Executor      string
	Schemabinding bool
	Definition    string
}

type Procedure struct {
	Id         string
	Connection string
	Name       string
	ObjectId   int64
	Schema     string
	Raw        string
	Properties ProcedureProps
}

func procedureFormatId(connectionId string, objectId int64) string {
	return fmt.Sprintf("%s/procedure/%d", connectionId, objectId)
}

func isProcedureId(id string) bool {
	return strings.Contains(id, "/procedure/")
}

func ParseProcedureId(ctx context.Context, id string) (procedure Procedure) {
	s := strings.Split(id, "/procedure/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /procedure/ exactly once")
		return
	}

	procedure.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	objectId, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse procedure id")
		return
	}

	procedure.ObjectId = objectId

	return
}

func CreateProcedureFromRaw(ctx context.Context, connection Connection, name string, schemaResourceId string, definition string) (procedure Procedure) {

	schema := GetSchemaFromId(ctx, connection, schemaResourceId, true)

	if logging.HasError(ctx) {
		return
	}

	regex := fmt.Sprintf("^[\\s]*create procedure %s.%s[\\s]*", schema.Name, name)
	if match, _ := regexp.MatchString("(?i)"+regex, definition); !match {
		logging.AddError(ctx, "Procedure creation failed", fmt.Sprintf("Procedure defintion should contain 'create procedure %s.%s'. The given definition was %s.", schema.Name, name, definition))
		return
	}

	_, err := connection.Connection.ExecContext(ctx, definition)

	if err != nil {
		logging.AddError(ctx, "Procedure creation failed", err)
		return
	}

	procedure = GetProcedureFromNameAndSchema(ctx, connection, name, schemaResourceId, false)
	if !logging.HasError(ctx) && procedure.Id == "" {
		logging.AddError(ctx, "Unable to read newly created procedure", fmt.Sprintf("Unable to read procedure %s after creation.", name))
	}

	return procedure
}

func buildProcedureQuery(name string, schemaName string, props ProcedureProps) string {

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
	if match, _ := regexp.MatchString("^[\\s]begin", strings.ToLower(definition)); !match {
		definition = fmt.Sprintf(`
BEGIN
 %s
END`, definition)
	}

	return fmt.Sprintf(`
create procedure %s.%s %s
%s%s
as 
%s
`, schemaName, name, arguments, execute_as, schemabinding, definition)
}

func CreateProcedureFromProperties(ctx context.Context, connection Connection, name string, schemaResourceId string, props ProcedureProps) (procedure Procedure) {

	schema := GetSchemaFromId(ctx, connection, schemaResourceId, true)

	if logging.HasError(ctx) {
		return
	}

	query := buildProcedureQuery(name, schema.Name, props)

	tflog.Info(ctx, query)

	return CreateProcedureFromRaw(ctx, connection, name, schemaResourceId, query)
}

func GetProcedureFromNameAndSchema(ctx context.Context, connection Connection, name string, schemaResourceId string, requiresExist bool) (procedure Procedure) {
	schema := ParseSchemaId(ctx, schemaResourceId)

	query := `
		select obj.object_id, mod.definition from sys.objects obj 
		inner join sys.sql_modules mod
		on obj.object_id = mod.object_id
		where obj.name = @name and obj.schema_id = @schema_id 
		and type in ('P')`

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("name", name), sql.Named("schema_id", schema.SchemaId)).Scan(&procedure.ObjectId, &procedure.Raw)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Procedure not found", fmt.Sprintf("Procedure with name %s doesn't exist", name))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading procedure %s failed", name), err)
		return
	}

	procedure.Id = procedureFormatId(connection.ConnectionId, procedure.ObjectId)
	procedure.Schema = schemaResourceId
	procedure.Name = name

	return
}

func GetProcedureFromObjectId(ctx context.Context, connection Connection, objectId int64, requiresExist bool) (procedure Procedure) {
	var schemaId int64
	query := `
		select obj.schema_id, obj.name, mod.definition from sys.objects obj 
		inner join sys.sql_modules mod
		on obj.object_id = mod.object_id
		where obj.object_id = @object_id
		and type in ('P')`

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("object_id", objectId)).Scan(&schemaId, &procedure.Name, &procedure.Raw)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Procedure not found", fmt.Sprintf("Procedure with objectId %d doesn't exist", objectId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading procedure %d failed", objectId), err)
		return
	}

	procedure.Id = procedureFormatId(connection.ConnectionId, objectId)
	procedure.ObjectId = objectId
	procedure.Schema = schemaFormatId(connection.ConnectionId, schemaId)
	procedure.Connection = connection.ConnectionId

	return
}

func GetProcedureFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (procedure Procedure) {
	procedure = ParseProcedureId(ctx, id)
	if logging.HasError(ctx) {
		return
	}

	if procedure.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	procedure = GetProcedureFromObjectId(ctx, connection, procedure.ObjectId, requiresExist)

	return procedure
}

func DropProcedure(ctx context.Context, connection Connection, id string) {

	procedure := GetProcedureFromId(ctx, connection, id, false)
	if logging.HasError(ctx) || procedure.Id == "" {
		return
	}
	schema := GetSchemaFromId(ctx, connection, procedure.Schema, true)
	if logging.HasError(ctx) {
		return
	}

	query := fmt.Sprintf("drop procedure %s.%s", schema.Name, procedure.Name)

	if _, err := connection.Connection.ExecContext(ctx, query); err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping procedure %s.%s failed", schema.Name, procedure.Name), err)
	}
}
