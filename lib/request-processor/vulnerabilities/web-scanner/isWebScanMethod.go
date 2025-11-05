package webscanner

import (
	"main/utils"
)

var methods = utils.StringSliceToMap([]string{"BADMETHOD", "BADHTTPMETHOD", "BADDATA", "BADMTHD", "BDMTHD"})

func isWebScanMethod(method string) bool {
	if _, ok := methods[method]; ok {
		return true
	}
	return false
}
