package sql_injection

import (
	"main/context"
	"main/log"
	"main/utils"
	zen_internals "main/vulnerabilities/zen-internals"
)

/**
 * This function goes over all the different input types in the context and checks
 * if it's a possible SQL Injection, if so the function returns an InterceptorResult
 */
func CheckContextForSqlInjection(sql string, operation string, dialect string) *utils.InterceptorResult {

	for _, source := range context.SOURCES {
		mapss := source.CacheGet()

		for str, path := range mapss {
			status, err := zen_internals.DetectSQLInjection(sql, str, utils.GetSqlDialectFromString(dialect))
			if err != nil {
				log.Error("Error while getting sql injection handler from zen internals", err)
				return nil
			}
			if status == 1 {
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
