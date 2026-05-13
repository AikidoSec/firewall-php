package main

import (
	"main/attack"
	"main/context"
	"main/instance"
	"main/log"
	idor "main/vulnerabilities/idor"
	sql_injection "main/vulnerabilities/sql-injection"
)

func OnPreSqlQueryExecuted(instance *instance.RequestProcessorInstance) string {
	query := context.GetSqlQuery(instance)
	dialect := context.GetSqlDialect(instance)
	operation := context.GetFunctionName(instance)
	if query == "" || dialect == "" {
		return ""
	}

	if context.IsEndpointProtectionTurnedOff(instance) {
		log.Infof(instance, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := sql_injection.CheckContextForSqlInjection(instance, query, operation, dialect)
	if res != nil {
		return attack.ReportAttackDetected(res, instance)
	}

	if context.GetIdorConfig(instance) != nil && !context.IsIdorDisabled(instance) {
		tenantId := context.GetTenantId(instance)
		sqlParams := context.GetSqlParams(instance)
		idorResult := idor.CheckForIdorViolation(instance, query, dialect, tenantId, sqlParams)
		if idorResult != "" {
			return idorResult
		}
	}

	return ""
}
