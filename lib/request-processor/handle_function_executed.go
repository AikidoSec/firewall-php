package main

import (
	. "main/aikido_types"
	"main/utils"
)

var functionExecutedHandlers = map[string]HandlerFunction{
	"curl_exec": OnFunctionExecutedCurl,

	"exec":       OnFunctionExecutedShell,
	"shell_exec": OnFunctionExecutedShell,
	"system":     OnFunctionExecutedShell,
	"passthru":   OnFunctionExecutedShell,
	"popen":      OnFunctionExecutedShell,
	"proc_open":  OnFunctionExecutedShell,

	// basename, chgrp, chmod, chown, clearstatcache, copy, dirname, disk_free_space ...
	"path_accessed": OnPathAccessed,
}

func OnFunctionExecuted(data map[string]interface{}) string {
	functionName := utils.MustGetFromMap[string](data, "function_name")
	parameters := utils.MustGetFromMap[map[string]interface{}](data, "parameters")

	utils.CheckIfKeyExists(functionExecutedHandlers, functionName)

	return functionExecutedHandlers[functionName](parameters)
}
