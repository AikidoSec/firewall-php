package sql_injection

import (
	"main/context"
	"main/helpers"
	"main/instance"
	"main/utils"
	zen_internals "main/vulnerabilities/zen-internals"
)

/**
 * This function goes over all the different input types in the context and checks
 * if it's a possible SQL Injection, if so the function returns an InterceptorResult
 */
func CheckContextForSqlInjection(instance *instance.RequestProcessorInstance, sql string, operation string, dialect string) *utils.InterceptorResult {
	trimmedSql := helpers.TrimInvisible(sql)
	dialectId := utils.GetSqlDialectFromString(dialect)

	blockInvalidSql := false
	if server := instance.GetCurrentServer(); server != nil {
		blockInvalidSql = server.AikidoConfig.BlockInvalidSql
	}

	for _, source := range context.SOURCES {
		mapss := source.CacheGet(instance)

		for str, path := range mapss {
			trimmedInputString := helpers.TrimInvisible(str)
			result := detectSQLInjection(trimmedSql, trimmedInputString, dialectId)

			if (result == zen_internals.SQLInjectionDetected) ||
				(result == zen_internals.SQLInjectionTokenizeFailed && blockInvalidSql) {
				metadata := map[string]string{
					"sql":     sql,
					"dialect": dialect,
				}
				if result == zen_internals.SQLInjectionTokenizeFailed {
					metadata["failedToTokenize"] = "true"
				}
				return &utils.InterceptorResult{
					Operation:     operation,
					Kind:          utils.Sql_injection,
					Source:        source.Name,
					PathToPayload: path,
					Metadata:      metadata,
					Payload:       str,
				}
			}
		}
	}
	return nil
}
