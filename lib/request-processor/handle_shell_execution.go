package main

import (
	"main/attack"
	"main/context"
	"main/instance"
	"main/log"
	shell_injection "main/vulnerabilities/shell-injection"
)

func OnPreShellExecuted(instance *instance.RequestProcessorInstance) string {
	cmd := context.GetCmd(instance)
	operation := context.GetFunctionName(instance)
	if cmd == "" {
		return ""
	}

	log.Info(instance, "Got shell command: ", cmd)

	if context.IsEndpointProtectionTurnedOff(instance) {
		log.Infof(instance, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := shell_injection.CheckContextForShellInjection(instance, cmd, operation)
	if res != nil {
		return attack.ReportAttackDetected(res, instance)
	}
	return ""
}
