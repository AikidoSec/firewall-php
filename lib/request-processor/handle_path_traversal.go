package main

import (
	. "main/aikido_types"
	"main/attack"
	"main/context"
	"main/log"
	"main/utils"
	path_traversal "main/vulnerabilities/path-traversal"
)

func OnPrePathAccessed() string {
	filename := utils.SanitizePath(context.GetFilename())
	filename2 := utils.SanitizePath(context.GetFilename2())
	operation := context.GetFunctionName()

	if filename == "" || operation == "" {
		return ""
	}

	if context.IsProtectionTurnedOff() {
		log.Infof("Protection is turned off -> will not run detection logic!")
		return ""
	}

	for _, f := range []string{filename, filename2} {
		res := context.CheckVulnerabilityOrGetFromCache(&FileAccessed{Filename: f, Operation: operation},
			path_traversal.CheckContextForPathTraversal,
			&context.Context.CachedFileAccessedResults)
		if res != nil {
			return attack.ReportAttackDetected(res)
		}
	}
	return ""
}
