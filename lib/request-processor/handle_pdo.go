package main

import (
	. "main/aikido_types"
	"main/attack"
	"main/context"
	"main/log"
	sql_injection "main/vulnerabilities/sql-injection"
)

func OnPreSqlQueryExecuted() string {
	query := context.GetSqlQuery()
	dialect := context.GetSqlDialect()
	operation := context.GetFunctionName()
	if query == "" || dialect == "" {
		return ""
	}
	log.Info("Got PDO query: ", query, " dialect: ", dialect)

	if context.IsProtectionTurnedOff() {
		log.Infof("Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := context.CheckVulnerabilityOrGetFromCache(&QueryExecuted{Query: query, Operation: operation, Dialect: dialect},
		sql_injection.CheckContextForSqlInjection,
		context.Context.CachedQueryExecutedResults)
	if res != nil {
		return attack.ReportAttackDetected(res)
	}
	return ""
}
