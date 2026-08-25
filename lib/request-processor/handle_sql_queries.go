package main

import (
	"main/attack"
	"main/context"
	"main/instance"
	"main/log"
	"main/utils"
	"main/vulnerabilities/idor"
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

	// Check SQL injection first: block malicious queries before spending time on IDOR analysis.
	res := sql_injection.CheckContextForSqlInjection(instance, query, operation, dialect)
	sqlInjectionAction := ""
	if res != nil {
		sqlInjectionAction = attack.ReportAttackDetected(res, instance)
		if utils.IsBlockingEnabled(instance.GetCurrentServer()) {
			return sqlInjectionAction
		}
	}

	if operation != "pg_prepare" && operation != "pg_send_prepare" {
		if idorMessage := idor.CheckContextForIdor(instance, query, dialect); idorMessage != "" {
			return idor.GetIdorThrowAction(idorMessage)
		}
	}

	return sqlInjectionAction
}
