package sql_injection

import (
	"main/context"
	"main/helpers"
	"main/utils"
)

/**
 * This function goes over all the different input types in the context and checks
 * if it's a possible SQL Injection, if so the function returns an InterceptorResult
 */
func CheckContextForSqlInjection(sql string, operation string, dialect string) *utils.InterceptorResult {
	trimmedSql := helpers.TrimInvisible(sql)
	dialectId := utils.GetSqlDialectFromString(dialect)

	for _, source := range context.SOURCES {
		mapss := source.CacheGet()

		for str, path := range mapss {
			trimmedInputString := helpers.TrimInvisible(str)
			if detectSQLInjection(trimmedSql, trimmedInputString, dialectId) {
				return &utils.InterceptorResult{
					Operation:     operation,
					Kind:          utils.Sql_injection,
					Source:        source.Name,
					PathToPayload: path,
					Metadata: map[string]string{
						"sql":     sql,
						"dialect": dialect,
					},
					Payload: str,
				}
			}
		}
	}
	return nil

}
