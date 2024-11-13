package shell_injection

import (
	. "main/aikido_types"
	"main/context"
	"main/utils"
)

func CheckContextForShellInjection(shellExecuted *ShellExecuted) *utils.InterceptorResult {
	for _, source := range context.SOURCES {
		mapss := source.CacheGet()

		for str, path := range mapss {
			if detectShellInjection(shellExecuted.Cmd, str) {
				return &utils.InterceptorResult{
					Operation:     shellExecuted.Operation,
					Kind:          utils.Shell_injection,
					Source:        source.Name,
					PathToPayload: path,
					Metadata: map[string]string{
						"command": shellExecuted.Cmd,
					},
					Payload: str,
				}
			}
		}
	}

	return nil
}
