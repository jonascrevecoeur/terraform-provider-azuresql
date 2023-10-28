package sql

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"terraform-provider-azuresql/internal/logging"
)

type SecurityPredicate struct {
	Id               string
	Connection       string
	SecurityPolicy   string
	PredicateId      int64
	Table            string
	Rule             string
	PredicateType    string
	BlockRestriction string
}

func securityPredicateFormatId(connectionId string, policyObjectId int64, predicateId int64) string {
	return fmt.Sprintf("%s/securitypredicate/%d/%d", connectionId, policyObjectId, predicateId)
}

func ParseSecurityPredicateId(ctx context.Context, id string) (securityPredicate SecurityPredicate) {
	s := strings.Split(id, "/securitypredicate/")

	if len(s) != 2 {
		logging.AddError(ctx, "ID format error", "id doesn't contain /securitypredicate/ exactly once")
		return
	}

	securityPredicate.Connection = s[0]

	s = strings.Split(s[1], "/")
	if len(s) != 2 {
		logging.AddError(ctx, "Invalid id", "Unable to parse policy predicate id")
		return
	}

	predicateId, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse function id")
		return
	}
	policyObjectId, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		logging.AddError(ctx, "Invalid id", "Unable to parse function id")
		return
	}

	securityPredicate.PredicateId = predicateId
	securityPredicate.SecurityPolicy = securityPolicyFormatId(securityPredicate.Connection, policyObjectId)

	return
}

func CreateSecurityPredicate(ctx context.Context, connection Connection, securityPolicyResourceId string, tableResourceId string,
	predicateType string, rule string, blockRestriction string) (securityPredicate SecurityPredicate) {

	policy := GetSecurityPolicyFromId(ctx, connection, securityPolicyResourceId, true)

	if logging.HasError(ctx) {
		return
	}

	schema := GetSchemaFromId(ctx, connection, policy.Schema, true)
	table := GetTableFromId(ctx, connection, tableResourceId, true)

	if logging.HasError(ctx) {
		return
	}

	tableSchema := GetSchemaFromId(ctx, connection, table.Schema, true)

	query := fmt.Sprintf(`
		ALTER SECURITY POLICY %s.%s  
    	ADD %s PREDICATE %s ON %s.%s %s`,
		schema.Name, policy.Name, predicateType, rule, tableSchema.Name, table.Name, blockRestriction)

	_, err := connection.Connection.ExecContext(ctx, query)

	if err != nil {
		logging.AddError(ctx, "Predicate creation failed", err)
		return
	}

	securityPredicate = GetSecurityPredicateFromSpec(ctx, connection, securityPolicyResourceId, tableResourceId,
		predicateType, false)
	if !logging.HasError(ctx) && securityPredicate.Id == "" {
		logging.AddError(ctx, "Unable to read newly created security predicatae", "Unable to read security predicate after creation.")
	}

	return securityPredicate
}

func GetSecurityPredicateFromSpec(ctx context.Context, connection Connection, securityPolicyResourceId string, tableResourceId string,
	predicateType string, requiresExist bool) (securityPredicate SecurityPredicate) {

	policy := ParseSecurityPolicyId(ctx, securityPolicyResourceId)
	table := parseTableId(ctx, tableResourceId)

	query := `
		SELECT security_predicate_id, predicate_definition, operation_desc FROM sys.security_predicates 
		where object_id = @policy_id and target_object_id = @table_id and predicate_type_desc = @type`

	var predicateId int64
	var definition string
	var operation sql.NullString

	err := connection.Connection.QueryRowContext(
		ctx, query, sql.Named("policy_id", policy.ObjectId), sql.Named("table_id", table.ObjectId),
		sql.Named("type", strings.ToUpper(predicateType))).Scan(&predicateId, &definition, &operation)
	switch {
	case err == sql.ErrNoRows:

		if requiresExist {
			logging.AddError(ctx, "Security predicate not found", fmt.Sprintf("Security predicate with doesn't exist"))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading security predicate failed"), err)
		return
	}

	return SecurityPredicate{
		Id:               securityPredicateFormatId(connection.ConnectionId, policy.ObjectId, predicateId),
		Connection:       connection.ConnectionId,
		SecurityPolicy:   securityPolicyResourceId,
		PredicateId:      predicateId,
		Table:            tableResourceId,
		Rule:             definition,
		PredicateType:    predicateType,
		BlockRestriction: strings.ToLower(parseNullString(operation)),
	}
}

func GetSecurityPredicateFromPolicyAndPredicateId(ctx context.Context, connection Connection, policyId int64, predicateId int64, requiresExist bool) (securityPredicate SecurityPredicate) {
	var tableId int64
	var definition, predicateType string
	var operation sql.NullString

	query := `
		SELECT target_object_id, predicate_definition, predicate_type_desc, operation_desc 
		FROM sys.security_predicates where security_predicate_id = @predicate_id and object_id = @policy_id`

	err := connection.Connection.QueryRowContext(ctx, query, sql.Named("predicate_id", predicateId), sql.Named("policy_id", policyId)).Scan(
		&tableId, &definition, &predicateType, &operation)
	switch {
	case err == sql.ErrNoRows:
		if requiresExist {
			logging.AddError(ctx, "Security predicate not found", fmt.Sprintf("Security predicate with policy/predicateId %d/%d doesn't exist", policyId, predicateId))
		}
		return
	case err != nil:
		logging.AddError(ctx, fmt.Sprintf("Reading security predicate %d/%d failed", policyId, predicateId), err)
		return
	}

	return SecurityPredicate{
		Id:               securityPredicateFormatId(connection.ConnectionId, policyId, predicateId),
		Connection:       connection.ConnectionId,
		SecurityPolicy:   securityPolicyFormatId(connection.ConnectionId, policyId),
		PredicateId:      predicateId,
		Table:            tableFormatId(connection.ConnectionId, tableId),
		Rule:             definition,
		PredicateType:    strings.ToLower(predicateType),
		BlockRestriction: strings.ToLower(parseNullString(operation)),
	}
}

func GetSecurityPredicateFromId(ctx context.Context, connection Connection, id string, requiresExist bool) (securityPredicate SecurityPredicate) {
	securityPredicate = ParseSecurityPredicateId(ctx, id)
	if logging.HasError(ctx) {
		return
	}
	policy := ParseSecurityPolicyId(ctx, securityPredicate.SecurityPolicy)

	if securityPredicate.Connection != connection.ConnectionId {
		logging.AddError(ctx, "Connection mismatch", fmt.Sprintf("Id %s doesn't belong to connection %s", id, connection.ConnectionId))
		return
	}

	securityPredicate = GetSecurityPredicateFromPolicyAndPredicateId(ctx, connection, policy.ObjectId, securityPredicate.PredicateId, requiresExist)

	return securityPredicate
}

func DropSecurityPredicate(ctx context.Context, connection Connection, id string) {

	predicate := GetSecurityPredicateFromId(ctx, connection, id, false)
	if logging.HasError(ctx) || predicate.Id == "" {
		return
	}
	policy := GetSecurityPolicyFromId(ctx, connection, predicate.SecurityPolicy, true)
	policySchema := GetSchemaFromId(ctx, connection, policy.Schema, true)
	if logging.HasError(ctx) {
		return
	}
	table := GetTableFromId(ctx, connection, predicate.Table, true)
	tableSchema := GetSchemaFromId(ctx, connection, table.Schema, true)
	if logging.HasError(ctx) {
		return
	}

	query := fmt.Sprintf(`
		alter security policy %s.%s drop %s predicate on %s.%s`,
		policySchema.Name, policy.Name, predicate.PredicateType, tableSchema.Name, table.Name)

	if _, err := connection.Connection.ExecContext(ctx, query); err != nil {
		logging.AddError(ctx, fmt.Sprintf("Dropping security predicate %d on %s.%s failed", predicate.PredicateId, policySchema.Name, policy.Name), err)
	}
}
