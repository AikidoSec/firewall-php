package main

import (
	"main/attack"
	"main/context"
	"main/log"
	path_traversal "main/vulnerabilities/path-traversal"
)

func OnPrePathAccessed() string {
	filename := context.GetFilename()
	filename2 := context.GetFilename2()
	operation := context.GetFunctionName()

	if filename == "" || operation == "" {
		return ""
	}

	if context.IsEndpointProtectionTurnedOff() {
		log.Infof("Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := path_traversal.CheckContextForPathTraversal(filename, operation, true)
	if res != nil {
		return attack.ReportAttackDetected(res)
	}

	if filename2 != "" {
		res = path_traversal.CheckContextForPathTraversal(filename2, operation, true)
		if res != nil {
			return attack.ReportAttackDetected(res)
		}
	}
	return ""
}
