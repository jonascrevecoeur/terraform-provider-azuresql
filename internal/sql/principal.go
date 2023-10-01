package sql

import (
	"context"
	"fmt"
	"terraform-provider-azuresql/internal/logging"
)

type Principal struct {
	Id          string
	Connection  string
	Name        string
	PrincipalId int64
	Type        string
}

func principalFormatId(connectionId string, id int64, principalType string) string {
	if principalType == "R" {
		return roleFormatId(connectionId, id)
	} else {
		return userFormatId(connectionId, id)
	}
}

func GetPrincipalFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (principal Principal) {
	if isRoleId(id) {
		role := GetRoleFromId(ctx, connection, id, requiresExist)
		if logging.HasError(ctx) {
			return
		}
		return Principal{
			Id:          id,
			Connection:  connection.ConnectionId,
			Name:        role.Name,
			PrincipalId: role.PrincipalId,
			Type:        "R",
		}
	} else if isUserId(id) {
		user := GetUserFromId(ctx, connection, id, requiresExist)
		if logging.HasError(ctx) {
			return
		}
		return Principal{
			Id:          id,
			Connection:  connection.ConnectionId,
			Name:        user.Name,
			PrincipalId: user.PrincipalId,
			Type:        user.Type,
		}
	} else {
		logging.AddError(ctx, "Invalid principal id", fmt.Sprintf("%s is not a valid user or role id", id))
		return
	}
}
