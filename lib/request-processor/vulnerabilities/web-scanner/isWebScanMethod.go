package webscanner

import "slices"

var methods = []string{"BADMETHOD", "BADHTTPMETHOD", "BADDATA", "BADMTHD", "BDMTHD"}

func isWebScanMethod(method string) bool {
	return slices.Contains(methods, method)
}
