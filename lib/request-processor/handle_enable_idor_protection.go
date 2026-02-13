package main

import (
	"encoding/json"
	"main/context"
	"main/log"
	idor "main/vulnerabilities/idor"
)

func OnEnableIdorProtection() string {
	tenantColumnName := context.GetIdorTenantColumnName()
	if tenantColumnName == "" {
		log.Warn("enable_idor_protection: tenant column name is empty!")
		return ""
	}

	excludedTablesJson := context.GetIdorExcludedTables()
	var excludedTables []string
	if excludedTablesJson != "" {
		if err := json.Unmarshal([]byte(excludedTablesJson), &excludedTables); err != nil {
			log.Warnf("enable_idor_protection: failed to parse excluded tables: %s", err)
			excludedTables = []string{}
		}
	}

	idor.SetIdorConfig(tenantColumnName, excludedTables)
	return ""
}
