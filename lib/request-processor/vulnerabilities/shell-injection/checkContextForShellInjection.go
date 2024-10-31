package shell_injection

import (
	"main/context"
	"main/log"
	"main/utils"
	zen_internals "main/vulnerabilities/zen-internals"
)

func CheckContextForShellInjection(command string, operation string) *utils.InterceptorResult {
	for _, source := range context.SOURCES {
		mapss := source.CacheGet()

		for str, path := range mapss {
			status, err := zen_internals.DetectShellInjection(command, str)
			if err != nil {
				log.Error("Error while getting shell injection handler from zen internals", err)
				return nil
			}
			if status == 1 {
				return &utils.InterceptorResult{
					Operation:     operation,
					Kind:          utils.Shell_injection,
					Source:        source.Name,
					PathToPayload: path,
					Metadata: map[string]string{
						"command": command,
					},
					Payload: str,
				}
			}
		}
	}

	return nil
}
