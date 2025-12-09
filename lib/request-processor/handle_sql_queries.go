package main

import (
	"main/attack"
	"main/context"
	"main/instance"
	"main/log"
	sql_injection "main/vulnerabilities/sql-injection"
)

func OnPreSqlQueryExecuted(inst *instance.RequestProcessorInstance) string {
	query := context.GetSqlQuery(inst)
	dialect := context.GetSqlDialect(inst)
	operation := context.GetFunctionName(inst)
	if query == "" || dialect == "" {
		return ""
	}

	if context.IsEndpointProtectionTurnedOff(inst) {
		log.Infof(inst, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := sql_injection.CheckContextForSqlInjection(inst, query, operation, dialect)
	if res != nil {
		return attack.ReportAttackDetected(res, inst)
	}
	return ""
}
