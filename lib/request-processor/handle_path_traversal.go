package main

import (
	"main/utils"
	path_traversal "main/vulnerabilities/path-traversal"
)

func OnPathAccessed(parameters map[string]interface{}) string {
	filename := utils.GetFromMap[string](parameters, "filename")
	filename2 := utils.GetFromMap[string](parameters, "filename2")
	operation := utils.GetFromMap[string](parameters, "operation")
	context := utils.ParseContext(utils.MustGetFromMap[map[string]interface{}](parameters, "context"))

	if filename == nil || operation == nil || context == nil {
		return "{}"
	}

	res := path_traversal.CheckContextForPathTraversal(*filename, *operation, context, true)
	if res != nil {
		return `{"action": "throw", "message": "Path traversal detected", "code": -1}`
	}

	if filename2 != nil && *filename2 != "" {
		res = path_traversal.CheckContextForPathTraversal(*filename2, *operation, context, true)
		if res != nil {
			return `{"action": "throw", "message": "Path traversal detected", "code": -1}`
		}
	}
	return "{}"
}
