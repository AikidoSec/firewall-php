package main

import (
	"fmt"
	"main/context"
	"main/globals"
	"main/log"
	"main/utils"
)

func OnRegisterParamMatcherEvent() string {
	param, regex := context.GetParamMatcher()
	if param == "" || regex == "" {
		return ""
	}

	if !utils.IsValidParamName(param) {
		return utils.GetMessageAction(fmt.Sprintf("Invalid param name: %s. Param names must match [a-zA-Z_]+", param))
	}

	server := globals.GetCurrentServer()
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
	log.Infof("Registered param matcher %s -> %s", param, regex)
	return ""
}
