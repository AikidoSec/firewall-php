package sql_injection

import (
	. "main/aikido_types"
	"main/context"
	"main/utils"
)

/**
 * This function goes over all the different input types in the context and checks
 * if it's a possible SQL Injection, if so the function returns an InterceptorResult
 */
func CheckContextForSqlInjection(queryExecuted *QueryExecuted) *utils.InterceptorResult {
	dialectId := utils.GetSqlDialectFromString(queryExecuted.Dialect)

	for _, source := range context.SOURCES {
		mapss := source.CacheGet()

		for str, path := range mapss {
			if detectSQLInjection(queryExecuted.Query, str, dialectId) {
				return &utils.InterceptorResult{
					Operation:     queryExecuted.Operation,
					Kind:          utils.Sql_injection,
					Source:        source.Name,
					PathToPayload: path,
					Metadata: map[string]string{
						"sql":     queryExecuted.Query,
						"dialect": queryExecuted.Dialect,
					},
					Payload: str,
				}
			}
		}
	}
	return nil

}
