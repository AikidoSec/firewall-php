package idor

import (
	"encoding/json"
	"fmt"
	"main/attack"
	"main/context"
	"main/log"
	"main/utils"
	zen_internals "main/vulnerabilities/zen-internals"
)

type TableRef struct {
	Name  string  `json:"name"`
	Alias *string `json:"alias,omitempty"`
}

type FilterColumn struct {
	Table             *string `json:"table,omitempty"`
	Column            string  `json:"column"`
	Value             string  `json:"value"`
	PlaceholderNumber *int    `json:"placeholder_number,omitempty"`
}

type InsertColumn struct {
	Column            string `json:"column"`
	Value             string `json:"value"`
	PlaceholderNumber *int   `json:"placeholder_number,omitempty"`
}

type SqlQueryResult struct {
	Kind          string           `json:"kind"`
	Tables        []TableRef       `json:"tables"`
	Filters       []FilterColumn   `json:"filters"`
	InsertColumns [][]InsertColumn `json:"insert_columns,omitempty"`
}

type AnalysisError struct {
	Error string `json:"error"`
}

func CheckForIdorViolation(query string, dialect string, tenantId string, sqlParams string) string {
	if context.GetIdorConfig() == nil {
		return ""
	}

	if tenantId == "" {
		return buildIdorViolationAction(
			"Zen IDOR protection: setTenantId() was not called for this request. Every request must have a tenant ID when IDOR protection is enabled.",
		)
	}

	dialectId := utils.GetSqlDialectFromString(dialect)
	resultJson := zen_internals.IdorAnalyzeSql(query, dialectId)
	if resultJson == "" {
		return buildIdorViolationAction("Zen IDOR protection: failed to analyze SQL query")
	}

	var analysisError AnalysisError
	if err := json.Unmarshal([]byte(resultJson), &analysisError); err == nil && analysisError.Error != "" {
		return buildIdorViolationAction(fmt.Sprintf("Zen IDOR protection: %s", analysisError.Error))
	}

	var queryResults []SqlQueryResult
	if err := json.Unmarshal([]byte(resultJson), &queryResults); err != nil {
		log.Warnf("Failed to parse IDOR analysis result: %s", err)
		return buildIdorViolationAction("Zen IDOR protection: failed to parse SQL analysis result")
	}

	params := parseSqlParams(sqlParams)

	for _, queryResult := range queryResults {
		switch queryResult.Kind {
		case "insert":
			if msg := checkInsert(queryResult, tenantId, params); msg != "" {
				return buildIdorViolationAction(msg)
			}
		case "select", "update", "delete":
			if msg := checkWhereFilters(queryResult, tenantId, params); msg != "" {
				return buildIdorViolationAction(msg)
			}
		}
		// DDL, transactions, etc. are allowed
	}

	return ""
}

func checkWhereFilters(queryResult SqlQueryResult, tenantId string, params *SqlParams) string {
	for _, table := range queryResult.Tables {
		if isExcludedTable(table.Name) {
			continue
		}

		tenantFilter := findTenantFilter(queryResult, table)
		if tenantFilter == nil {
			return fmt.Sprintf(
				"Zen IDOR protection: query on table '%s' is missing a filter on column '%s'",
				table.Name, context.GetIdorConfig().TenantColumnName,
			)
		}

		resolvedValue := resolveValue(tenantFilter.Value, tenantFilter.PlaceholderNumber, params)
		if resolvedValue != "" && resolvedValue != tenantId {
			return fmt.Sprintf(
				"Zen IDOR protection: query on table '%s' filters '%s' with value '%s' but tenant ID is '%s'",
				table.Name, context.GetIdorConfig().TenantColumnName, resolvedValue, tenantId,
			)
		}
	}

	return ""
}

func checkInsert(queryResult SqlQueryResult, tenantId string, params *SqlParams) string {
	for _, table := range queryResult.Tables {
		if isExcludedTable(table.Name) {
			continue
		}

		if queryResult.InsertColumns == nil {
			// INSERT ... SELECT without explicit columns
			return fmt.Sprintf(
				"Zen IDOR protection: INSERT on table '%s' is missing column '%s'",
				table.Name, context.GetIdorConfig().TenantColumnName,
			)
		}

		for _, row := range queryResult.InsertColumns {
			tenantCol := findTenantColumn(row)
			if tenantCol == nil {
				return fmt.Sprintf(
					"Zen IDOR protection: INSERT on table '%s' is missing column '%s'",
					table.Name, context.GetIdorConfig().TenantColumnName,
				)
			}

			resolvedValue := resolveValue(tenantCol.Value, tenantCol.PlaceholderNumber, params)
			if resolvedValue != "" && resolvedValue != tenantId {
				return fmt.Sprintf(
					"Zen IDOR protection: INSERT on table '%s' sets '%s' to '%s' but tenant ID is '%s'",
					table.Name, context.GetIdorConfig().TenantColumnName, resolvedValue, tenantId,
				)
			}
		}
	}

	return ""
}

func findTenantFilter(queryResult SqlQueryResult, table TableRef) *FilterColumn {
	for i, f := range queryResult.Filters {
		if f.Column != context.GetIdorConfig().TenantColumnName {
			continue
		}

		if f.Table != nil {
			// Qualified column (e.g. u.tenant_id) — match against table name or alias
			if *f.Table == table.Name || (table.Alias != nil && *f.Table == *table.Alias) {
				return &queryResult.Filters[i]
			}
		} else {
			// Unqualified column — only safe to match in single-table queries
			if len(queryResult.Tables) == 1 {
				return &queryResult.Filters[i]
			}
		}
	}

	return nil
}

func findTenantColumn(row []InsertColumn) *InsertColumn {
	for i, c := range row {
		if c.Column == context.GetIdorConfig().TenantColumnName {
			return &row[i]
		}
	}
	return nil
}

func isExcludedTable(tableName string) bool {
	for _, excluded := range context.GetIdorConfig().ExcludedTables {
		if excluded == tableName {
			return true
		}
	}
	return false
}

type SqlParams struct {
	Positional []interface{}
	Named      map[string]string
}

func parseSqlParams(sqlParamsJson string) *SqlParams {
	if sqlParamsJson == "" {
		return nil
	}

	var positional []interface{}
	if err := json.Unmarshal([]byte(sqlParamsJson), &positional); err == nil {
		return &SqlParams{Positional: positional}
	}

	var named map[string]string
	if err := json.Unmarshal([]byte(sqlParamsJson), &named); err == nil {
		return &SqlParams{Named: named}
	}

	return nil
}

// Returns empty string for unresolvable placeholders (skips value validation).
func resolveValue(value string, placeholderNumber *int, params *SqlParams) string {
	if params != nil {
		if placeholderNumber != nil && params.Positional != nil {
			idx := *placeholderNumber
			if idx >= 0 && idx < len(params.Positional) {
				if str, ok := params.Positional[idx].(string); ok {
					return str
				}
				return ""
			}
			return ""
		}

		if params.Named != nil && len(value) > 0 && value[0] == ':' {
			if resolved, ok := params.Named[value]; ok {
				return resolved
			}
			return ""
		}
	}

	if placeholderNumber != nil {
		return ""
	}
	if len(value) > 0 && (value[0] == '$' || value[0] == ':' || value == "?") {
		return ""
	}
	return value
}

func buildIdorViolationAction(message string) string {
	actionMap := map[string]interface{}{
		"action":         "throw",
		"message":        message,
		"code":           500,
		"idor_violation": true,
	}
	actionJson, err := json.Marshal(actionMap)
	if err != nil {
		return attack.GetThrowAction(message, 500)
	}
	return string(actionJson)
}
