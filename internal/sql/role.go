package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type Role struct {
	Id          string
	Connection  string
	Name        string
	PrincipalId int64
	Owner       string
}

func roleFormatId(connectionId string, rolePrincipalId int64) string {
	return fmt.Sprintf("%s/role/%d", connectionId, rolePrincipalId)
}

func isRoleId(id string) (isRole bool) {
	return strings.Contains(id, "/role/")
}

func ParseRoleId(ctx context.Context, id string) (role Role) {
	s := strings.Split(id, "/role/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /role/ exactly once")
		return
	}

	role.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	principal_id, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", fmt.Sprintf("Unable to parse role id %s", id))
		return
	}

	role.PrincipalId = principal_id

	return
}

func roleOwnerFormatId(ctx context.Context, connectionId string, principalId int64, ownerType string) (owner string) {
	if ownerType == "R" {
		return roleFormatId(connectionId, principalId)
	} else {
		return userFormatId(connectionId, principalId)
	}
}

func CreateRole(ctx context.Context, connection Connection, name string, owner string) (role Role) {
	tflog.Info(ctx, fmt.Sprintf("Creating role %s", name))

	query := fmt.Sprintf("create role [%s]", name)

	if owner != "" {
		var owner_name string
		if isRoleId(owner) {
			owner_role := GetRoleFromId(ctx, connection, owner, true)
			if logging.HasError(ctx) {
				return
			}
			owner_name = owner_role.Name
		} else if isUserId(owner) {
			user := GetUserFromId(ctx, connection, owner, true)
			if logging.HasError(ctx) {
				return
			}
			owner_name = user.Name
		} else {
			logging.AddError(ctx, "Invalid owner id", fmt.Sprintf("%s is not a valid user or role id", owner))
			return
		}
		query += fmt.Sprintf(" authorization [%s]", owner_name)
	}

	_, err := connection.Connection.ExecContext(ctx, query)

	logging.AddError(ctx, fmt.Sprintf("Role creation failed for role %s", name), err)

	// set requiresExist to false in order to specify a custom error message
	role = GetRoleFromName(ctx, connection, name, false)
	if !logging.HasError(ctx) && role.Id == "" {
		logging.AddError(ctx, "Unable to read newly created role", fmt.Sprintf("Unable to read role %s after creation.", name))
	}

	return role
}

func GetRoleFromName(ctx context.Context, connection Connection, name string, requiresExist bool) (role Role) {

	var id, ownerId int64
	var ownerType string

	query := `
		select role.principal_id, owner.principal_id as owner_id, owner.type as owner_type from sys.database_principals role
		left join sys.database_principals owner on owner.principal_id = role.owning_principal_id
		where role.name = @name and role.type = 'R'`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("name", name)).
		Scan(&id, &ownerId, &ownerType))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Role not found", fmt.Sprintf("Role with name %s doesn't exist", name))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading role %s failed", name), err)
		return
	}

	return Role{
		Id:          roleFormatId(connection.ConnectionId, id),
		Connection:  connection.ConnectionId,
		Name:        name,
		PrincipalId: id,
		Owner:       roleOwnerFormatId(ctx, connection.ConnectionId, ownerId, ownerType),
	}

}

func GetRoleFromPrincipalId(ctx context.Context, connection Connection, principalId int64, requiresExist bool) (role Role) {

	var ownerId int64
	var name, ownerType string

	query := `
		select role.name, owner.principal_id as owner_id, owner.type as owner_type from sys.database_principals role
		left join sys.database_principals owner on owner.principal_id = role.owning_principal_id
		where role.principal_id = @id and role.type = 'R'
		`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("id", principalId)).
		Scan(&name, &ownerId, &ownerType))

	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Role not found", fmt.Sprintf("Role with principal id %d doesn't exist", principalId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading role %s failed", name), err)
		return
	}

	return Role{
		Id:          roleFormatId(connection.ConnectionId, principalId),
		Connection:  connection.ConnectionId,
		Name:        name,
		PrincipalId: principalId,
		Owner:       roleOwnerFormatId(ctx, connection.ConnectionId, ownerId, ownerType),
	}

}

// Get user from the azuresql terraform id
// requiresExist: Raise an error when the user doesn't exist
func GetRoleFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (role Role) {
	role = ParseRoleId(ctx, id)
	if logging.HasError(ctx) {
		return
	}

	if role.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	role = GetRoleFromPrincipalId(ctx, connection, role.PrincipalId, requiresExist)

	return role
}

func UpdateRoleName(ctx context.Context, connection Connection, id string, name string) {
	role := GetRoleFromId(ctx, connection, id, true)
	if logging.HasError(ctx) {
		return
	}

	query := fmt.Sprintf("alter role %s with name = %s", role.Name, name)
	_, err := connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Alter name for role %s failed", role.Name), err)
	}
}

func UpdateRoleOwner(ctx context.Context, connection Connection, id string, owner string) {
	role := GetRoleFromId(ctx, connection, id, true)
	if logging.HasError(ctx) {
		return
	}

	var owner_name string
	if isRoleId(owner) {
		owner_role := GetRoleFromId(ctx, connection, owner, true)
		if logging.HasError(ctx) {
			return
		}
		owner_name = owner_role.Name
	} else if isUserId(owner) {
		user := GetUserFromId(ctx, connection, owner, true)
		if logging.HasError(ctx) {
			return
		}
		owner_name = user.Name
	} else {
		logging.AddError(ctx, "Invalid owner id", fmt.Sprintf("%s is not a valid user or role id", owner))
		return
	}

	query := fmt.Sprintf("alter authorization on role::%s to %s", role.Name, owner_name)
	_, err := connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Alter owner for role %s failed", role.Name), err)
	}
}

func DropRole(ctx context.Context, connection Connection, principalId int64) {

	tflog.Info(ctx, fmt.Sprintf("Dropping role %d", principalId))

	role := GetRoleFromPrincipalId(ctx, connection, principalId, false)
	if logging.HasError(ctx) || role.Id == "" {
		return
	}

	query := fmt.Sprintf(`
		IF EXISTS (SELECT 1 FROM sys.database_principals WHERE name = '%[1]s')
		BEGIN
			DROP ROLE [%[1]s];
		END;
		`, role.Name)
	var err error
	_, err = connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping role %s failed", role.Name), err)
	}
}
