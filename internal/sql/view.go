package sql

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type View struct {
	Id            string
	Connection    string
	Name          string
	ObjectId      int64
	Schema        string
	Schemabinding bool
	CheckOption   bool
	Definition    string
}

func viewFormatId(connectionId string, objectId int64) string {
	return fmt.Sprintf("%s/view/%d", connectionId, objectId)
}

func isViewId(id string) bool {
	return strings.Contains(id, "/view/")
}

func ParseViewId(ctx context.Context, id string) (view View) {
	s := strings.Split(id, "/view/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /view/ exactly once")
		return
	}

	view.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	objectId, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse view id")
		return
	}

	view.ObjectId = objectId

	return
}

func CreateViewFromDefinition(ctx context.Context, connection Connection, name string, schemaResourceId string, definition string, schemabinding bool, checkOption bool) (view View) {

	schema := GetSchemaFromId(ctx, connection, schemaResourceId, true)

	if logging.HasError(ctx) {
		return
	}

	arg_schemabinding := ""
	if schemabinding {
		arg_schemabinding = "with schemabinding"
	}

	arg_checkoption := ""
	if checkOption {
		arg_checkoption = "with check option"
	}

	query := fmt.Sprintf(`
create view %s.%s %s as (
	%s
)
%s
	`, schema.Name, name, arg_schemabinding, definition, arg_checkoption)

	_, err := connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, "View creation failed", err)
		return
	}

	view = GetViewFromNameAndSchema(ctx, connection, name, schemaResourceId, false)
	if !logging.HasError(ctx) && view.Id == "" {
		logging.AddError(ctx, "Unable to read newly created view", fmt.Sprintf("Unable to read view %s after creation.", name))
	}

	return view
}

func cleanDefinition(definition string) string {
	retval := strings.TrimSpace(definition)
	retval = strings.TrimSuffix(retval, "with check option")
	retval = strings.TrimSpace(retval)
	if retval[0] == '(' && retval[len(retval)-1] == ')' {
		retval = strings.TrimSpace(retval[2:(len(retval) - 1)])
	}
	return retval
}
func extractViewDefintion(ctx context.Context, statement string) (defintion string) {
	// use regexp for case insensitive split --> important when importing
	re := regexp.MustCompile(`(?i)\s+as\s+`)

	split := re.Split(statement, 2)

	if len(split) != 2 {
		logging.AddError(ctx, "Failed to parse view definition", fmt.Sprintf("Couldn't find ` as ` indicating the start of the definition %s", statement))
		return
	}

	return cleanDefinition(split[1])
}

func extractViewCheckOption(ctx context.Context, statement string) bool {
	value, _ := regexp.MatchString("(?i)with check option", statement)
	return value
}

func IsViewDefinitionEquivalent(ctx context.Context, definition1 string, definition2 string) bool {
	if definition1 == "" || definition2 == "" {
		return (definition1 == "" && definition2 == "")
	}

	return cleanDefinition(definition1) == cleanDefinition(definition2)
}

func GetViewFromNameAndSchema(ctx context.Context, connection Connection, name string, schemaResourceId string, requiresExist bool) (view View) {
	var statement string
	var schemabinding bool
	schema := ParseSchemaId(ctx, schemaResourceId)

	query := `
		select obj.object_id, mod.definition, mod.is_schema_bound from sys.objects obj
		inner join sys.sql_modules mod
		on obj.object_id = mod.object_id
		where obj.name = @name and obj.schema_id = @schema_id
		and type = 'V'`

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("name", name), sql.Named("schema_id", schema.SchemaId)).Scan(&view.ObjectId, &statement, &schemabinding)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "View not found", fmt.Sprintf("View with name %s doesn't exist", name))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading view %s failed", name), err)
		return
	}

	view.Id = viewFormatId(connection.ConnectionId, view.ObjectId)
	view.Schema = schemaResourceId
	view.Name = name
	view.Schemabinding = schemabinding
	view.Definition = extractViewDefintion(ctx, statement)
	view.CheckOption = extractViewCheckOption(ctx, statement)

	return
}

func GetViewFromObjectId(ctx context.Context, connection Connection, objectId int64, requiresExist bool) (view View) {
	var statement string
	var schemaId int64
	var schemabinding bool

	query := `
		select obj.schema_id, obj.name, mod.definition, mod.is_schema_bound from sys.objects obj
		inner join sys.sql_modules mod
		on obj.object_id = mod.object_id
		where obj.object_id = @object_id
		and type = 'V'`

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("object_id", objectId)).Scan(&schemaId, &view.Name, &statement, &schemabinding)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "View not found", fmt.Sprintf("View with objectId %d doesn't exist", objectId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading view %d failed", objectId), err)
		return
	}

	view.Id = viewFormatId(connection.ConnectionId, objectId)
	view.ObjectId = objectId
	view.Schema = schemaFormatId(connection.ConnectionId, schemaId)
	view.Connection = connection.ConnectionId
	view.Definition = extractViewDefintion(ctx, statement)
	view.Schemabinding = schemabinding
	view.CheckOption = extractViewCheckOption(ctx, statement)

	return
}

func GetViewFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (view View) {
	view = ParseViewId(ctx, id)
	if logging.HasError(ctx) {
		return
	}

	if view.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	view = GetViewFromObjectId(ctx, connection, view.ObjectId, requiresExist)

	return view
}

func DropView(ctx context.Context, connection Connection, id string) {

	view := GetViewFromId(ctx, connection, id, false)
	if logging.HasError(ctx) || view.Id == "" {
		return
	}
	schema := GetSchemaFromId(ctx, connection, view.Schema, true)
	if logging.HasError(ctx) {
		return
	}

	query := fmt.Sprintf("drop view %s.%s", schema.Name, view.Name)

	if _, err := connection.Connection.ExecContext(ctx, query); err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping view %s.%s failed", schema.Name, view.Name), err)
	}
}
