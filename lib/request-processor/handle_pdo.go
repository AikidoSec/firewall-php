package main

import (
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

	if context.IsEndpointProtectionTurnedOff() {
		log.Infof("Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := sql_injection.CheckContextForSqlInjection(query, operation, dialect)
	if res != nil {
		return attack.ReportAttackDetected(res)
	}
	return ""
}
