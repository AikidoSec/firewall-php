package shell_injection

import (
	"main/context"
	"main/helpers"
	"main/utils"
)

func CheckContextForShellInjection(command string, operation string) *utils.InterceptorResult {
	trimmedCommand := helpers.TrimInvisible(command)
	for _, source := range context.SOURCES {
		mapss := source.CacheGet()

		for str, path := range mapss {
			trimmedInputString := helpers.TrimInvisible(str)
			if detectShellInjection(trimmedCommand, trimmedInputString) {
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
