package main

import (
	"main/attack"
	"main/context"
	"main/instance"
	"main/log"
	path_traversal "main/vulnerabilities/path-traversal"
)

func OnPrePathAccessed(instance *instance.RequestProcessorInstance) string {
	filename := context.GetFilename(instance)
	filename2 := context.GetFilename2(instance)
	operation := context.GetFunctionName(instance)

	if filename == "" || operation == "" {
		return ""
	}

	if context.IsEndpointProtectionTurnedOff(instance) {
		log.Infof(instance, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := path_traversal.CheckContextForPathTraversal(instance, filename, operation, true)
	if res != nil {
		return attack.ReportAttackDetected(res, instance)
	}

	if filename2 != "" {
		res = path_traversal.CheckContextForPathTraversal(instance, filename2, operation, true)
		if res != nil {
			return attack.ReportAttackDetected(res, instance)
		}
	}
	return ""
}
