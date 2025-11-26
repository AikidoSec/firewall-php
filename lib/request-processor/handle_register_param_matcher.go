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
	log.Infof("Got param matcher name: %s", param)
	log.Infof("Got param matcher regex: %s", regex)

	server := globals.GetCurrentServer()
	if server == nil {
		return ""
	}

	if _, exists := server.ParamMatchers[param]; exists {
		log.Debugf("Param matcher %s already exists, skipping", param)
		return ""
	}

	regexCompiled, err := utils.CompileCustomPattern(regex)
	if err != nil {
		errorMessage := fmt.Sprintf("Error compiling param matcher %s -> regex \"%s\": %s", param, regex, err.Error())
		log.Info(errorMessage)
		return utils.GetMessageAction(errorMessage)
	}
	server.ParamMatchers[param] = regexCompiled
	log.Infof("Registered param matcher %s -> %s", param, regex)
	return ""
}
