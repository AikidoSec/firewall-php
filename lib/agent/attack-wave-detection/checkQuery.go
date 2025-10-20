package attackwavedetection

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

func checkQuery(queryParams map[string]string) bool {

	for _, param := range queryParams {
		// Performance optimization
		if len(param) < 5 || len(param) > 1000 {
			continue
		}

		upperParam := strings.ToUpper(param)
		for _, keyword := range keywords {
			if strings.Contains(upperParam, keyword) {
				return true
			}
		}
	}
	return false

}
