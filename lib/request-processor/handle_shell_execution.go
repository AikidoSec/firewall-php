package main

import (
	. "main/aikido_types"
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

	if context.IsProtectionTurnedOff() {
		log.Infof("Protection is turned off -> will not run detection logic!")
		return ""
	}
	res := context.CheckVulnerabilityOrGetFromCache(&ShellExecuted{Cmd: cmd, Operation: operation},
		shell_injection.CheckContextForShellInjection,
		&context.Context.CachedShellExecutedResults)
	if res != nil {
		return attack.ReportAttackDetected(res)
	}
	return ""
}
