package main

import (
	"main/attack"
	"main/context"
	"main/instance"
	"main/log"
	path_traversal "main/vulnerabilities/path-traversal"
)

func OnPrePathAccessed(inst *instance.RequestProcessorInstance) string {
	filename := context.GetFilename(inst)
	filename2 := context.GetFilename2(inst)
	operation := context.GetFunctionName(inst)

	if filename == "" || operation == "" {
		return ""
	}

	if context.IsEndpointProtectionTurnedOff(inst) {
		log.Infof(inst, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := path_traversal.CheckContextForPathTraversal(inst, filename, operation, true)
	if res != nil {
		return attack.ReportAttackDetected(res, inst)
	}

	if filename2 != "" {
		res = path_traversal.CheckContextForPathTraversal(inst, filename2, operation, true)
		if res != nil {
			return attack.ReportAttackDetected(res, inst)
		}
	}
	return ""
}
