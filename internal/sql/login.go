package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type Login struct {
	Id         string
	Connection string
	Name       string
	Password   string
	Sid        string
}

func loginFormatId(connectionId string, name string, sid string) string {
	return fmt.Sprintf("%s/login/%s/%s", connectionId, name, sid)
}

// retrieve name and sid from a tf login id
func ParseLoginId(ctx context.Context, id string) (login Login) {
	s := strings.Split(id, "/login/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /login/ exactly once")
		return
	}

	login.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	s = strings.Split(s[1], "/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", fmt.Sprintf("Invalid login id %s", id))
		return
	}

	login.Name = s[0]
	login.Sid = s[1]

	return
}

func CreateLogin(ctx context.Context, connection Connection, name string) (login Login) {
	password := generatePassword(20, 3, 4, 5)

	query := fmt.Sprintf("create login %s with password = '%s'", name, password)

	_, err := connection.Connection.ExecContext(ctx, query)
	logging.AddError(ctx, fmt.Sprintf("Login creation failed for login %s", name), err)

	login = GetLoginFromName(ctx, connection, name)
	if logging.HasError(ctx) || login.Id == "" {
		logging.AddError(ctx, "Resource creation failed", fmt.Sprintf("Unable to read login %s after creation.", name))
	}

	login.Password = password

	return login
}

func GetLoginFromName(ctx context.Context, connection Connection, name string) (login Login) {

	var sid string
	// sid is stored as varbinary, convert returns the hexadecimal representation as a string
	// this is the most usefull go representation for performing future queries
	query := fmt.Sprintf("select convert(varchar(max), sid, 1) as sid from sys.sql_logins where name = '%s'", name)

	err := (connection.
		Connection.
		QueryRowContext(ctx, query).
		Scan(&sid))

	switch {
	case err == sql.ErrNoRows:
		// login doesn't exist
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading login %s failed", name), err)
		return
	}

	return Login{
		Id:         loginFormatId(connection.ConnectionId, name, sid),
		Connection: connection.ConnectionId,
		Name:       name,
		Sid:        sid,
	}
}

func GetLoginFromSid(ctx context.Context, connection Connection, sid string) (login Login) {
	var name string
	query := fmt.Sprintf("select name from sys.sql_logins where sid = %s", sid)

	err := (connection.
		Connection.
		QueryRowContext(ctx, query).
		Scan(&name))

	switch {
	case err == sql.ErrNoRows:
		// login doesn't exist
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading login %s failed", sid), err)
		return
	}

	return Login{
		Id:         loginFormatId(connection.ConnectionId, name, sid),
		Connection: connection.ConnectionId,
		Name:       name,
		Sid:        sid,
	}
}

func DropLogin(ctx context.Context, connection Connection, sid string) {

	login := GetLoginFromSid(ctx, connection, sid)

	if logging.HasError(ctx) || login.Id == "" {
		return
	}

	query := fmt.Sprintf("drop login [%s]", login.Name)
	var err error
	_, err = connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping login %s failed", login.Name), err)
	}
}
