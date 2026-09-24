package idor

import (
	"encoding/json"
	"fmt"
	"main/context"
	"main/instance"
	"main/utils"
	zen_internals "main/vulnerabilities/zen-internals"
	"strconv"
	"strings"
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
	IsPlaceholder     bool    `json:"is_placeholder"`
}

type InsertColumn struct {
	Column            string `json:"column"`
	Value             string `json:"value"`
	PlaceholderNumber *int   `json:"placeholder_number,omitempty"`
	IsPlaceholder     bool   `json:"is_placeholder"`
}

type SqlQueryResult struct {
	Kind          string            `json:"kind"`
	Tables        []TableRef        `json:"tables"`
	Filters       []FilterColumn    `json:"filters"`
	InsertColumns *[][]InsertColumn `json:"insert_columns,omitempty"`
}

type IdorConfig struct {
	Enabled          bool     `json:"enabled"`
	TenantColumnName string   `json:"tenantColumnName"`
	ExcludedTables   []string `json:"excludedTables"`
}

func getIdorConfig(instance *instance.RequestProcessorInstance) IdorConfig {
	var config IdorConfig
	if raw := context.GetIdorConfigJson(instance); raw != "" {
		json.Unmarshal([]byte(raw), &config)
	}
	return config
}

// analyzeSql parses (dialect, query) via zen-internals; cached by query text.
func analyzeSql(query string, dialect int) ([]SqlQueryResult, bool) {
	cacheKey := strconv.Itoa(dialect) + ":" + query
	if cached, ok := sqlAnalysisCache.get(cacheKey); ok {
		return cached, true
	}

	raw := zen_internals.IdorAnalyzeSql(query, dialect)
	if raw == "" {
		return nil, false
	}

	var results []SqlQueryResult
	if err := json.Unmarshal([]byte(raw), &results); err != nil {
		return nil, false
	}

	sqlAnalysisCache.set(cacheKey, results)
	return results, true
}

// lookupNamedParam tries both spellings of a named placeholder, since PDO accepts
// "tenant_id" or ":tenant_id" as an execute() key but zen-internals reports it as written.
func lookupNamedParam(params map[string]string, name string) (string, bool) {
	if v, ok := params[name]; ok {
		return v, true
	}
	if len(name) > 1 {
		switch name[0] {
		case ':', '@', '$':
			if v, ok := params[name[1:]]; ok {
				return v, true
			}
		default:
			if v, ok := params[":"+name]; ok {
				return v, true
			}
		}
	}
	return "", false
}

// resolveValue returns the literal value for a filter/insert column, resolving
// placeholders (":name", "?", "$1", ...) against the bound params passed to execute().
func resolveValue(value string, placeholderNumber *int, isPlaceholder bool, params map[string]string) (string, bool) {
	if !isPlaceholder {
		return value, true
	}
	if v, ok := lookupNamedParam(params, value); ok {
		return v, true
	}
	// placeholder_number is 0-based, matching how the PHP extension keys positional params.
	if placeholderNumber != nil {
		if v, ok := params[strconv.Itoa(*placeholderNumber)]; ok {
			return v, true
		}
	}
	return "", false
}

func joinWithLimit(items []string, limit int) string {
	if len(items) <= limit {
		return strings.Join(items, ", ")
	}
	return strings.Join(items[:limit], ", ") + ", ..."
}

func tenantNotSetMessage(tables []string) string {
	noun := "table"
	if len(tables) > 1 {
		noun = "tables"
	}
	return fmt.Sprintf("Zen IDOR protection: query on %s '%s' requires a tenant ID, but set_tenant_id() was not called.", noun, joinWithLimit(tables, 5))
}

func nonExcludedTables(result SqlQueryResult, excludedTables map[string]bool) []string {
	var tables []string
	for _, table := range result.Tables {
		if !excludedTables[table.Name] {
			tables = append(tables, table.Name)
		}
	}
	return tables
}

func checkWhereFilters(result SqlQueryResult, tenantColumn string, tenantId string, excludedTables map[string]bool, params map[string]string) string {
	for _, table := range result.Tables {
		if excludedTables[table.Name] {
			continue
		}

		var matched *FilterColumn
		for i := range result.Filters {
			f := &result.Filters[i]
			if f.Column != tenantColumn {
				continue
			}
			if f.Table != nil {
				if *f.Table == table.Name || (table.Alias != nil && *f.Table == *table.Alias) {
					matched = f
					break
				}
				continue
			}
			// Unqualified filter: only trust it when the query has a single table.
			if len(result.Tables) == 1 {
				matched = f
				break
			}
		}

		if matched == nil {
			return fmt.Sprintf("Zen IDOR protection: query on table '%s' is missing a filter on column '%s'", table.Name, tenantColumn)
		}

		value, resolved := resolveValue(matched.Value, matched.PlaceholderNumber, matched.IsPlaceholder, params)
		if !resolved {
			return fmt.Sprintf("Zen IDOR protection: query on table '%s' has a placeholder for '%s' that could not be resolved", table.Name, tenantColumn)
		}
		if value != tenantId {
			return fmt.Sprintf("Zen IDOR protection: query on table '%s' filters '%s' with value '%s' but tenant ID is '%s'", table.Name, tenantColumn, value, tenantId)
		}
	}
	return ""
}

func checkInsert(result SqlQueryResult, tenantColumn string, tenantId string, excludedTables map[string]bool, params map[string]string) string {
	for _, table := range result.Tables {
		if excludedTables[table.Name] {
			continue
		}

		// INSERT ... SELECT without explicit columns: we can't verify the tenant column.
		if result.InsertColumns == nil {
			return fmt.Sprintf("Zen IDOR protection: INSERT on table '%s' is missing column '%s'", table.Name, tenantColumn)
		}

		for _, row := range *result.InsertColumns {
			var matched *InsertColumn
			for i := range row {
				if row[i].Column == tenantColumn {
					matched = &row[i]
					break
				}
			}
			if matched == nil {
				return fmt.Sprintf("Zen IDOR protection: INSERT on table '%s' is missing column '%s'", table.Name, tenantColumn)
			}

			value, resolved := resolveValue(matched.Value, matched.PlaceholderNumber, matched.IsPlaceholder, params)
			if !resolved {
				return fmt.Sprintf("Zen IDOR protection: INSERT on table '%s' has a placeholder for '%s' that could not be resolved", table.Name, tenantColumn)
			}
			if value != tenantId {
				return fmt.Sprintf("Zen IDOR protection: INSERT on table '%s' sets '%s' to '%s' but tenant ID is '%s'", table.Name, tenantColumn, value, tenantId)
			}
		}
	}
	return ""
}

// CheckContextForIdor returns a violation message, or "" if the query is allowed.
func CheckContextForIdor(instance *instance.RequestProcessorInstance, query string, dialect string) string {
	config := getIdorConfig(instance)
	if !config.Enabled || config.TenantColumnName == "" {
		return ""
	}
	if context.GetIdorIgnored(instance) {
		return ""
	}

	tenantId := context.GetTenantId(instance)

	tenantColumn := config.TenantColumnName
	excludedTables := map[string]bool{}
	for _, table := range config.ExcludedTables {
		excludedTables[table] = true
	}

	params := map[string]string{}
	if raw := context.GetSqlParams(instance); raw != "" {
		json.Unmarshal([]byte(raw), &params)
	}

	dialectId := utils.GetSqlDialectFromString(dialect)
	results, ok := analyzeSql(query, dialectId)
	if !ok {
		return ""
	}

	for _, result := range results {
		switch result.Kind {
		case "select", "update", "delete", "insert":
		default:
			// DDL, transactions and session commands never need a tenant.
			continue
		}

		// Only require a tenant once a checked table is touched, so migrations, health
		// checks, and the pre-auth lookup that resolves the tenant itself still work.
		if tenantId == "" {
			if tables := nonExcludedTables(result, excludedTables); len(tables) > 0 {
				return tenantNotSetMessage(tables)
			}
			continue
		}

		switch result.Kind {
		case "select", "update", "delete":
			if msg := checkWhereFilters(result, tenantColumn, tenantId, excludedTables, params); msg != "" {
				return msg
			}
		case "insert":
			if msg := checkInsert(result, tenantColumn, tenantId, excludedTables, params); msg != "" {
				return msg
			}
		}
	}
	return ""
}

// GetIdorThrowAction builds the "throw" action JSON, marked "idorViolation" so
// the PHP extension throws even when AIKIDO_BLOCK is off (see docs/idor-protection.md).
func GetIdorThrowAction(message string) string {
	actionMap := map[string]interface{}{
		"action":        "throw",
		"message":       message,
		"code":          500,
		"idorViolation": true,
	}
	actionJson, err := json.Marshal(actionMap)
	if err != nil {
		return ""
	}
	return string(actionJson)
}
