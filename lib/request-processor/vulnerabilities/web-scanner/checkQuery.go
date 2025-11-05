package webscanner

import "strings"

var keywords = []string{
	"SELECT (CASE WHEN",
	"SELECT COUNT(",
	"SLEEP(",
	"WAITFOR DELAY",
	"SELECT LIKE(CHAR(",
	"INFORMATION_SCHEMA.COLUMNS",
	"INFORMATION_SCHEMA.TABLES",
	"MD5(",
	"DBMS_PIPE.RECEIVE_MESSAGE",
	"SYSIBM.SYSTABLES",
	"RANDOMBLOB(",
	"SELECT * FROM",
	"1'='1",
	"PG_SLEEP(",
	"UNION ALL SELECT",
	"../",
}

func checkQuery(queryParams map[string]interface{}) bool {
	if queryParams == nil {
		return false
	}
	for _, param := range queryParams {
		// Performance optimization
		paramString, ok := param.(string)
		if !ok {
			continue
		}
		if len(paramString) < 5 || len(paramString) > 1000 {
			continue
		}

		upperParam := strings.ToUpper(param.(string))
		for _, keyword := range keywords {
			if strings.Contains(upperParam, keyword) {
				return true
			}
		}
	}
	return false

}
