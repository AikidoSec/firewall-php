package main

import (
	"main/attack"
	"main/context"
	"main/instance"
	"main/log"
	shell_injection "main/vulnerabilities/shell-injection"
)

func OnPreShellExecuted(inst *instance.RequestProcessorInstance) string {
	cmd := context.GetCmd(inst)
	operation := context.GetFunctionName(inst)
	if cmd == "" {
		return ""
	}

	log.Info(inst, "Got shell command: ", cmd)

	if context.IsEndpointProtectionTurnedOff(inst) {
		log.Infof(inst, "Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := shell_injection.CheckContextForShellInjection(inst, cmd, operation)
	if res != nil {
		return attack.ReportAttackDetected(res, inst)
	}
	return ""
}
