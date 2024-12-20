package main

import (
	"main/attack"
	"main/context"
	"main/log"
	shell_injection "main/vulnerabilities/shell-injection"
)

func OnPreShellExecuted() string {
	cmd := context.GetCmd()
	operation := context.GetFunctionName()
	if cmd == "" {
		return ""
	}

	log.Info("Got shell command: ", cmd)

	if context.IsEndpointProtectionTurnedOff() {
		log.Infof("Protection is turned off -> will not run detection logic!")
		return ""
	}

	res := shell_injection.CheckContextForShellInjection(cmd, operation)
	if res != nil {
		return attack.ReportAttackDetected(res)
	}
	return ""
}
