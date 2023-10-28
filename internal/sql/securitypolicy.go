package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type SecurityPolicy struct {
	Id         string
	Connection string
	Name       string
	ObjectId   int64
	Schema     string
}

func securityPolicyFormatId(connectionId string, objectId int64) string {
	return fmt.Sprintf("%s/securitypolicy/%d", connectionId, objectId)
}

func ParseSecurityPolicyId(ctx context.Context, id string) (securityPolicy SecurityPolicy) {
	s := strings.Split(id, "/securitypolicy/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /securitypolicy/ exactly once")
		return
	}

	securityPolicy.Connection = s[0]

	if logging.HasError(ctx) {
		return
	}

	objectId, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse policy id")
		return
	}

	securityPolicy.ObjectId = objectId

	return
}

func CreateSecurityPolicy(ctx context.Context, connection Connection, name string, schemaResourceId string) (securityPolicy SecurityPolicy) {

	schema := GetSchemaFromId(ctx, connection, schemaResourceId, true)

	if logging.HasError(ctx) {
		return
	}

	query := fmt.Sprintf("CREATE SECURITY POLICY %s.%s", schema.Name, name)

	_, err := connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, "Security policy creation failed", err)
		return
	}

	securityPolicy = GetSecurityPolicyFromNameAndSchema(ctx, connection, name, schemaResourceId, false)
	if !logging.HasError(ctx) && securityPolicy.Id == "" {
		logging.AddError(ctx, "Unable to read newly created security policy", fmt.Sprintf("Unable to read security policy %s after creation.", name))
	}

	return securityPolicy
}

func GetSecurityPolicyFromNameAndSchema(ctx context.Context, connection Connection, name string, schemaResourceId string, requiresExist bool) (securityPolicy SecurityPolicy) {
	schema := ParseSchemaId(ctx, schemaResourceId)

	query := "SELECT object_id FROM sys.security_policies where name = @name and schema_id = @schema_id"

	var objectId int64

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("name", name), sql.Named("schema_id", schema.SchemaId)).Scan(&objectId)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Security policy not found", fmt.Sprintf("Security policy with name %s doesn't exist", name))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading security policy %s failed", name), err)
		return
	}

	return SecurityPolicy{
		Id:         securityPolicyFormatId(connection.ConnectionId, objectId),
		Connection: connection.ConnectionId,
		Name:       name,
		ObjectId:   objectId,
		Schema:     schemaResourceId,
	}
}

func GetSecurityPolicyFromObjectId(ctx context.Context, connection Connection, objectId int64, requiresExist bool) (securityPolicy SecurityPolicy) {
	var schemaId int64
	var name string
	query := "SELECT name, schema_id FROM sys.security_policies where object_id = @object_id"

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("object_id", objectId)).Scan(&name, &schemaId)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Security policy not found", fmt.Sprintf("Security policy with objectId %d doesn't exist", objectId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading security policy %d failed", objectId), err)
		return
	}

	return SecurityPolicy{
		Id:         securityPolicyFormatId(connection.ConnectionId, objectId),
		Connection: connection.ConnectionId,
		Name:       name,
		ObjectId:   objectId,
		Schema:     schemaFormatId(connection.ConnectionId, schemaId),
	}
}

func GetSecurityPolicyFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (securityPolicy SecurityPolicy) {
	securityPolicy = ParseSecurityPolicyId(ctx, id)
	if logging.HasError(ctx) {
		return
	}

	if securityPolicy.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	securityPolicy = GetSecurityPolicyFromObjectId(ctx, connection, securityPolicy.ObjectId, requiresExist)

	return securityPolicy
}

func DropSecurityPolicy(ctx context.Context, connection Connection, id string) {

	policy := GetSecurityPolicyFromId(ctx, connection, id, false)
	if logging.HasError(ctx) || policy.Id == "" {
		return
	}
	schema := GetSchemaFromId(ctx, connection, policy.Schema, true)
	if logging.HasError(ctx) {
		return
	}

	query := fmt.Sprintf("drop security policy %s.%s", schema.Name, policy.Name)

	if _, err := connection.Connection.ExecContext(ctx, query); err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping security policy %s.%s failed", schema.Name, policy.Name), err)
	}
}
