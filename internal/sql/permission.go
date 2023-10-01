package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type Scope struct {
	ResourceType string
	Name         string
	Id           int64
}

type Permission struct {
	Id          string
	Connection  string
	Scope       string
	ScopeId     int64
	Principal   string
	PrincipalId int64
	Permission  string
	ScopeType   string
}

func permissionFormatId(connectionId string, principalId int64, permission string, permissionType string, targetId int64) string {
	return fmt.Sprintf("%s/permission/%d/%s/%s/%d", connectionId, principalId, permission, permissionType, targetId)
}

func parsePermissionId(ctx context.Context, id string) (permission Permission) {
	s := strings.Split(id, "/permission/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /permission/ exactly once")
		return
	}

	permission.Connection = s[0]

	s = strings.Split(s[1], "/")
	if len(s) != 4 {
		logging.AddError(ctx, "Invalid id", "Unable to parse permission id")
		return
	}

	var principal_id, scope_id int64
	var err error

	principal_id, err = strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse permission id")
		return
	}

	scope_id, err = strconv.ParseInt(s[3], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse permission id")
		return
	}

	permission.ScopeId = scope_id
	permission.PrincipalId = principal_id
	permission.Permission = s[1]
	permission.ScopeType = s[2]
	return
}

func objectFormatId(ctx context.Context, connectionId string, objectId int64, objectType string) string {
	if objectType == "U" {
		return tableFormatId(connectionId, objectId)
	}

	logging.AddError(ctx, "Unrecognized object type", fmt.Sprintf("Unexpected object type %s found", objectType))
	return ""
}

func scopeFormatId(ctx context.Context, connection Connection, scopeId int64, scopeType string) string {
	if scopeType == "schema" {
		return schemaFormatId(connection.ConnectionId, scopeId)
	}
	if scopeType == "object" {
		var objectType string
		query := "select type from sys.objects where object_id = @object_id"

		err := (connection.
			Connection.
			QueryRowContext(ctx, query, sql.Named("object_id", scopeId)).
			Scan(&objectType))

		switch {
		case err == sql.ErrNoRows:

			logging.AddError(ctx, "Object not found",
				fmt.Sprintf("Object with id %d doesn't exist",
					scopeId))

			return ""
		case err != nil:
			logging.AddError(ctx, fmt.Sprintf("Reading object %d failed", scopeId), err)
			return ""
		}

		return objectFormatId(ctx, connection.ConnectionId, scopeId, objectType)

	}
	logging.AddError(ctx, "Invalid scope", fmt.Sprintf("scopeFormatId not (yet) implemented for resources of type %s", scopeType))
	return ""
}

func GetScopeFromId(ctx context.Context, connection Connection, scopeResourceId string, requiresExist bool) (scope Scope) {
	if isSchemaId(scopeResourceId) {
		schema := GetSchemaFromId(ctx, connection, scopeResourceId, requiresExist)
		if schema.Id == "" {
			return
		}
		return Scope{
			ResourceType: "schema",
			Name:         schema.Name,
			Id:           schema.SchemaId,
		}
	}
	if isTableId(scopeResourceId) {
		table := GetTableFromId(ctx, connection, scopeResourceId, requiresExist)
		if table.Id == "" {
			return
		}
		return Scope{
			ResourceType: "object",
			Name:         fmt.Sprintf("%s.%s", table.SchemaName, table.Name),
			Id:           table.ObjectId,
		}
	}
	logging.AddError(ctx, "Invalid scope", fmt.Sprintf("Scope %s is not valid", scopeResourceId))
	return Scope{}
}

func CreatePermission(ctx context.Context, connection Connection, scopeResourceId string, principalResourceId string, permissionName string) (permission Permission) {

	principal := GetPrincipalFromId(ctx, connection, principalResourceId, true)
	if logging.HasError(ctx) {
		return
	}

	scope := GetScopeFromId(ctx, connection, scopeResourceId, true)
	if logging.HasError(ctx) {
		return
	}

	var query string
	if scope.ResourceType == "schema" {
		query = fmt.Sprintf("grant %s on schema::%s to [%s]", permissionName, scope.Name, principal.Name)
	} else if scope.ResourceType == "object" {
		query = fmt.Sprintf("grant %s on object::%s to [%s]", permissionName, scope.Name, principal.Name)
	}

	_, err := connection.Connection.ExecContext(ctx, query)
	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Failed to grant permission %s on %s to %s", permissionName, scope.Name, principal.Name), err)
		return
	}

	return Permission{
		Id:          permissionFormatId(connection.ConnectionId, principal.PrincipalId, permissionName, scope.ResourceType, scope.Id),
		Connection:  connection.ConnectionId,
		Scope:       scopeResourceId,
		ScopeId:     scope.Id,
		Principal:   principalResourceId,
		PrincipalId: principal.PrincipalId,
		Permission:  permissionName,
		ScopeType:   scope.ResourceType,
	}
}

func GetAllPermissions(ctx context.Context, connection Connection, scopeResourceId string, principalResourceId string) (permissions []string) {

	principal := GetPrincipalFromId(ctx, connection, principalResourceId, true)
	if logging.HasError(ctx) {
		return
	}

	scope := GetScopeFromId(ctx, connection, scopeResourceId, true)
	if logging.HasError(ctx) {
		return
	}

	query := `
		select permission_name from sys.database_permissions 
		where major_id = @scope_id and grantee_principal_id=@principal_id 
		and state = 'G'`

	rows, err := connection.Connection.QueryContext(ctx, query, sql.Named("scope_id", scope.Id),
		sql.Named("principal_id", principal.PrincipalId))
	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Failed to retrieve permissions for %s on %s", scope.Name, principal.Name), err)
		return
	}

	for rows.Next() {
		var permission string
		if err := rows.Scan(&permission); err != nil {
			// Check for a scan error.
			// Query rows will be closed with defer.
			logging.AddError(ctx, fmt.Sprintf("Failed to retrieve permissions for %s on %s", scope.Name, principal.Name), err)
			return
		}
		permissions = append(permissions, permission)
	}
	return
}

func GetPermissionFromId(ctx context.Context, connection Connection, permissionResourceId string, requiresExist bool) (permission Permission) {

	permission = parsePermissionId(ctx, permissionResourceId)

	if logging.HasError(ctx) {
		return
	}

	if permission.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", permissionResourceId, connection.ConnectionId))
		return
	}

	// check existence
	var principalType string
	query := `
		select principals.type from sys.database_permissions permissions
		left join sys.database_principals principals 
		on permissions.grantee_principal_id = principals.principal_id
		where permissions.major_id = @scope_id and permissions.grantee_principal_id=@principal_id
		and state = 'G'
		`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("scope_id", permission.ScopeId), sql.Named("principal_id", permission.PrincipalId)).
		Scan(&principalType))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Permission not found",
				fmt.Sprintf("Permission %s for principal id %d on resource %d doesn't exist",
					permission.Permission, permission.PrincipalId, permission.ScopeId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading permission %s for principal %d failed", permission.Permission, permission.PrincipalId), err)
		return
	}

	return Permission{
		Id:          permissionResourceId,
		Connection:  connection.ConnectionId,
		Scope:       scopeFormatId(ctx, connection, permission.ScopeId, permission.ScopeType),
		ScopeId:     permission.ScopeId,
		Principal:   principalFormatId(connection.ConnectionId, permission.PrincipalId, principalType),
		PrincipalId: permission.PrincipalId,
		Permission:  permission.Permission,
		ScopeType:   permission.ScopeType,
	}
}

func DropPermission(ctx context.Context, connection Connection, scopeResourceId string, principalResourceId string, permissionName string) {

	principal := GetPrincipalFromId(ctx, connection, principalResourceId, false)
	if logging.HasError(ctx) || principal.Id == "" {
		return
	}

	scope := GetScopeFromId(ctx, connection, scopeResourceId, false)
	if logging.HasError(ctx) || scope.Name == "" {
		return
	}

	var query string
	if scope.ResourceType == "schema" {
		query = fmt.Sprintf("revoke %s on schema::%s to [%s]", permissionName, scope.Name, principal.Name)
	} else if scope.ResourceType == "object" {
		query = fmt.Sprintf("revoke %s on object::%s to [%s]", permissionName, scope.Name, principal.Name)
	}

	_, err := connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Failed to revoke permission %s on %s for %s", permissionName, scope.Name, principal.Name), err)
		return
	}
}
