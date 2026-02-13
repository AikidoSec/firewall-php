package main

import (
	"main/attack"
	"main/context"
	"main/log"
	idor "main/vulnerabilities/idor"
	sql_injection "main/vulnerabilities/sql-injection"
)

func OnPreSqlQueryExecuted() string {
	query := context.GetSqlQuery()
	dialect := context.GetSqlDialect()
	operation := context.GetFunctionName()
	if query == "" || dialect == "" {
		return ""
	}

	if context.IsEndpointProtectionTurnedOff() {
		log.Infof("Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := sql_injection.CheckContextForSqlInjection(query, operation, dialect)
	if res != nil {
		return attack.ReportAttackDetected(res)
	}

	if idor.IsIdorEnabled() && !context.IsIdorDisabled() {
		tenantId := context.GetTenantId()
		sqlParams := context.GetSqlParams()
		idorResult := idor.CheckForIdorViolation(query, dialect, tenantId, sqlParams)
		if idorResult != "" {
			return idorResult
		}
	}

	return ""
}
