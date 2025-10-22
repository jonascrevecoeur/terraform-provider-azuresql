package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type User struct {
	Id             string
	Connection     string
	Name           string
	PrincipalId    int64
	Type           string
	Authentication string
	Login          string
	Password       string
	Sid            string
}

func userFormatId(connectionId string, userPrincipalId int64) string {
	return fmt.Sprintf("%s/user/%d", connectionId, userPrincipalId)
}

func isUserId(id string) (isRole bool) {
	//isRole, _ = regexp.MatchString("^[^/]*/role/[^/]*$", "/role/")
	return strings.Contains(id, "/user/")
}

func ParseUserId(ctx context.Context, id string) (user User) {
	s := strings.Split(id, "/user/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /user/ exactly once")
		return
	}

	user.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	principal_id, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse user id")
		return
	}

	user.PrincipalId = principal_id

	return
}

func describeUserType(ctx context.Context, userType string) (userTypeLong string) {
	switch userType {
	case "S":
		return "SQL user"
	case "X":
		return "AD group"
	case "E":
		return "AD user"
	default:
		logging.AddWarning(ctx, "Unrecognized user type", fmt.Sprintf("Unrecognized user type %s", userType))
		return "Unrecongized"
	}
}

func describeAuthentication(ctx context.Context, authentication_type int64) (authentication string) {
	switch authentication_type {
	case 0:
		return "WithoutLogin"
	case 4:
		return "AzureAD"
	case 1:
		return "SQLLogin"
	case 2:
		return "DBSQLLogin"
	default:
		logging.AddError(ctx, "Unrecognized authentication type", fmt.Sprintf("Unrecognized authentication type %d", authentication_type))
		return "Unrecongized"
	}
}

func CreateUser(ctx context.Context, connection Connection, name string, password string, authentication string, loginId string, entraid_identifier string) (user User) {

	query := fmt.Sprintf("create user [%s]", name)

	if authentication == "AzureAD" {
		if entraid_identifier == "" {
			query += " from external provider"
		} else {
			sid := ObjectIDToDatabaseSID(ctx, entraid_identifier)
			if logging.HasError(ctx) {
				return
			}
			query += " with sid=" + sid + ", type=E"
		}
	} else if authentication == "SQLLogin" {
		login := ParseLoginId(ctx, loginId)
		login_connection := ParseConnectionId(ctx, login.Connection)

		if login_connection.Provider != connection.Provider || login_connection.Server != connection.Server {
			logging.AddError(ctx, "Login/user incompatible",
				fmt.Sprintf("Login from %s is incompatible with a user from %s", login_connection.ConnectionId, connection.ConnectionId))
			return
		}

		query += fmt.Sprintf(" from login [%s]", login.Name)
	} else if authentication == "DBSQLLogin" {
		query += fmt.Sprintf(" with password='%s'", password)
	} else if authentication == "WithoutLogin" {
		query += " without login"
	}

	if connection.Provider == "fabric" {
		/*
			Fabric doesn't have a `create user` statement, since users are managed at the workspace level
			The user appears in the sys.database_principal table the first time a transaction is performed.
			The query below triggers this by granting and revoking access to a user.
		*/
		query = fmt.Sprintf(`
BEGIN TRANSACTION;
GRANT CONNECT to [%s]
REVOKE CONNECT to [%s]
COMMIT;		`, name, name)
	}

	_, err := connection.Connection.ExecContext(ctx, query)

	logging.AddError(ctx, fmt.Sprintf("User creation failed for user %s", name), err)

	user = GetUserFromName(ctx, connection, name)
	if user.Id == "" && !logging.HasError(ctx) {
		logging.AddError(ctx, "Unable to read newly created user", fmt.Sprintf("Unable to read user %s after creation.", name))
	}

	return user
}

func GetUserFromName(ctx context.Context, connection Connection, name string) (user User) {

	var id, authentication_type int64
	var userType, sid string

	query := `
		select principal_id, type, authentication_type, sid
		from sys.database_principals
		where name = @name and type != 'R'
		`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("name", name)).
		Scan(&id, &userType, &authentication_type, &sid))

	switch {
	case err == sql.ErrNoRows:
		// user doesn't exist
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading user %s failed", name), err)
		return
	}

	return User{
		Id:             userFormatId(connection.ConnectionId, id),
		Connection:     connection.ConnectionId,
		Name:           name,
		PrincipalId:    id,
		Type:           describeUserType(ctx, userType),
		Authentication: describeAuthentication(ctx, authentication_type),
		Sid:            sid,
	}

}

// Get user from the azuresql terraform id
// requiresExist: Raise an error when the user doesn't exist
func GetUserFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (user User) {
	user = ParseUserId(ctx, id)
	if logging.HasError(ctx) {
		return
	}

	if user.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	user = GetUserFromPrincipalId(ctx, connection, user.PrincipalId)

	if requiresExist && user.Id == "" {
		logging.AddError(ctx, "User not found", fmt.Sprintf("User with id %s doesn't exist", id))
		return
	}

	return user
}

func GetUserFromPrincipalId(ctx context.Context, connection Connection, principalId int64) (user User) {

	var name, userType, sid string
	var authentication_type int64

	query := `
		select name, type, authentication_type, sid
		from sys.database_principals
		where principal_id = @id and type != 'R'
		`

	err := (connection.
		Connection.
		QueryRowContext(ctx, query, sql.Named("id", principalId)).
		Scan(&name, &userType, &authentication_type, &sid))

	switch {
	case err == sql.ErrNoRows:
		// user doesn't exist
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading user %s failed", name), err)
		return
	}

	return User{
		Id:             userFormatId(connection.ConnectionId, principalId),
		Connection:     connection.ConnectionId,
		Name:           name,
		PrincipalId:    principalId,
		Type:           describeUserType(ctx, userType),
		Authentication: describeAuthentication(ctx, authentication_type),
		Sid:            sid,
	}

}

func DropUser(ctx context.Context, connection Connection, principalId int64) {

	user := GetUserFromPrincipalId(ctx, connection, principalId)
	if logging.HasError(ctx) || user.Id == "" {
		return
	}

	query := fmt.Sprintf(`
		IF EXISTS (SELECT 1 FROM sys.database_principals WHERE name = N'%[1]s' AND type IN ('S','E','X','U'))
		BEGIN
			DROP USER [%[1]s];
		END;
		`, user.Name)
	var err error
	_, err = connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping user %s failed", user.Name), err)
	}
}
