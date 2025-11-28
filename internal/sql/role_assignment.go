package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type RoleAssignment struct {
	Id              string
	Connection      string
	Role            string
	RolePrincipalId int64
	Principal       string
	PrincipalId     int64
}

func roleAssignmentFormatId(connectionId string, rolePrincipalId int64, principalId int64) string {
	return fmt.Sprintf("%s/roleassignment/%d/%d", connectionId, rolePrincipalId, principalId)
}

func ParseRoleAssignmentId(ctx context.Context, id string) (assignment RoleAssignment) {
	s := strings.Split(id, "/roleassignment/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /roleassignment/ exactly once")
		return
	}

	assignment.Connection = s[0]

	s = strings.Split(s[1], "/")
	if len(s) != 2 {
		logging.AddError(ctx, "Invalid id", "Unable to parse role assignment id")
		return
	}

	var role_id, principal_id int64
	var err error

	role_id, err = strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse role assignment id")
		return
	}

	principal_id, err = strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse role assignment id")
		return
	}

	assignment.PrincipalId = principal_id
	assignment.RolePrincipalId = role_id
	return
}

func CreateRoleAssignment(ctx context.Context, connection Connection, roleResourceId string, principalResourceId string) (assignment RoleAssignment) {

	principal := GetPrincipalFromId(ctx, connection, principalResourceId, true)
	if logging.HasError(ctx) {
		return
	}

	role := GetRoleFromId(ctx, connection, roleResourceId, true)
	if logging.HasError(ctx) {
		return
	}

	var err error
	if connection.Provider == "synapsededicated" {
		_, err = connection.Connection.ExecContext(ctx,
			"EXEC sp_addrolemember @rolename, @membername",
			sql.Named("rolename", role.Name),
			sql.Named("membername", principal.Name),
		)
	} else if connection.IsServerConnection {
		query := fmt.Sprintf("alter server role [%s] add member [%s]", role.Name, principal.Name)
		_, err = connection.Connection.ExecContext(ctx, query)
	} else {
		query := fmt.Sprintf("alter role [%s] add member [%s]", role.Name, principal.Name)
		_, err = connection.Connection.ExecContext(ctx, query)
	}
	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Failed to assign role %s %s", role.Name, principal.Name), err)
		return
	}

	return RoleAssignment{
		Id:              roleAssignmentFormatId(connection.ConnectionId, role.PrincipalId, principal.PrincipalId),
		Connection:      connection.ConnectionId,
		Role:            roleResourceId,
		RolePrincipalId: role.PrincipalId,
		Principal:       principalResourceId,
		PrincipalId:     principal.PrincipalId,
	}
}

func GetRoleAssignmentFromId(ctx context.Context, connection Connection, roleAssignmentResourceId string, requiresExist bool) (roleAssignment RoleAssignment) {

	roleAssignment = ParseRoleAssignmentId(ctx, roleAssignmentResourceId)

	if logging.HasError(ctx) {
		return
	}

	if roleAssignment.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", roleAssignmentResourceId, connection.ConnectionId))
		return
	}

	// check existence
	var query string
	if connection.IsServerConnection {
		query = `
			select principals.type from sys.server_role_members role
			left join sys.database_principals principals on role.member_principal_id = principals.principal_id
			where role.role_principal_id = @role_id and role.member_principal_id = @principal_id
			`
	} else {
		query = `
			select principals.type from sys.database_role_members role
			left join sys.database_principals principals on role.member_principal_id = principals.principal_id
			where role.role_principal_id = @role_id and role.member_principal_id = @principal_id
			`
	}

	var principalType string

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("role_id", roleAssignment.RolePrincipalId), sql.Named("principal_id", roleAssignment.PrincipalId)).
		Scan(&principalType))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Role assignment not found",
				fmt.Sprintf("Role assignment principal id %d on role %d doesn't exist",
					roleAssignment.PrincipalId, roleAssignment.RolePrincipalId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading roleAssignment failed"), err)
		return
	}

	return RoleAssignment{
		Id:              roleAssignmentResourceId,
		Connection:      connection.ConnectionId,
		Role:            roleFormatId(connection.ConnectionId, roleAssignment.RolePrincipalId),
		RolePrincipalId: roleAssignment.RolePrincipalId,
		Principal:       principalFormatId(connection.ConnectionId, roleAssignment.PrincipalId, principalType),
		PrincipalId:     roleAssignment.PrincipalId,
	}
}

func DropRoleAssignment(ctx context.Context, connection Connection, roleAssignmentResourceId string) {

	assignment := GetRoleAssignmentFromId(ctx, connection, roleAssignmentResourceId, true)
	if logging.HasError(ctx) || assignment.Id == "" {
		return
	}

	role := GetRoleFromId(ctx, connection, assignment.Role, true)
	if logging.HasError(ctx) {
		return
	}

	principal := GetPrincipalFromId(ctx, connection, assignment.Principal, true)
	if logging.HasError(ctx) {
		return
	}

	var err error
	if connection.Provider == "synapsededicated" {
		_, err = connection.Connection.ExecContext(ctx,
			"EXEC sp_droprolemember @rolename, @membername",
			sql.Named("rolename", role.Name),
			sql.Named("membername", principal.Name),
		)
	} else if connection.IsServerConnection {
		query := fmt.Sprintf("alter server role [%s] drop member [%s]", role.Name, principal.Name)
		_, err = connection.Connection.ExecContext(ctx, query)
	} else {
		query := fmt.Sprintf("alter role [%s] drop member [%s]", role.Name, principal.Name)
		_, err = connection.Connection.ExecContext(ctx, query)
	}
	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Failed to remove %s from role %s", principal.Name, role.Name), err)
		return
	}
}
