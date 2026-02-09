package main

import (
	"fmt"
	"main/context"
	"main/instance"
	"main/log"
	"main/utils"
)

func OnRegisterParamMatcherEvent(instance *instance.RequestProcessorInstance) string {
	param, regex := context.GetParamMatcher(instance)
	if param == "" || regex == "" {
		return ""
	}

	if !utils.IsValidParamName(param) {
		return utils.GetMessageAction(fmt.Sprintf("Invalid param name: %s. Param names must match [a-zA-Z_]+", param))
	}

	server := instance.GetCurrentServer()
	if server == nil {
		return ""
	}

	if _, exists := server.ParamMatchers[param]; exists {
		return ""
	}

	regexCompiled, err := utils.CompileCustomPattern(regex)
	if err != nil {
		return utils.GetMessageAction(fmt.Sprintf("Error compiling param matcher %s -> regex \"%s\": %s", param, regex, err.Error()))
	}
	server.ParamMatchers[param] = regexCompiled
	log.Infof(instance, "Registered param matcher %s -> %s", param, regex)
	return ""
}
