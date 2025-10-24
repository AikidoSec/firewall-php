package webscanner

func IsWebScanner(method string, route string, queryParams map[string]interface{}) bool {
	if method != "" && isWebScanMethod(method) {
		return true
	}

	if route != "" && isWebScanPath(route) {
		return true
	}

	if queryParams != nil && checkQuery(queryParams) {
		return true
	}

	return false
}
