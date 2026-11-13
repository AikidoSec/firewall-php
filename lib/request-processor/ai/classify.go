package ai

import (
	"main/helpers"
	"main/ipc/protos"
	"net/url"
	"strings"
)

type ResponseKind uint8

const (
	unknownResponse ResponseKind = iota
	openAIChat
	openAIResponses
	anthropic
	gemini
	bedrockConverse
	bedrockInvoke
)

func ClassifyURL(rawURL string) (ResponseKind, *protos.AIUsage) {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return unknownResponse, nil
	}

	path := u.EscapedPath()
	var kind ResponseKind
	var model string
	if strings.HasSuffix(path, "/chat/completions") {
		kind = openAIChat
	} else if strings.HasSuffix(path, "/responses") {
		kind = openAIResponses
	} else if strings.HasSuffix(path, "/v1/messages") ||
		modelInPath(path, "/publishers/anthropic/models/", ":rawPredict") != "" {
		kind = anthropic
	} else if pathModel := modelInPath(path, "/models/", ":generateContent"); pathModel != "" && !strings.Contains(pathModel, ":") {
		kind = gemini
		model = pathModel
	} else if pathModel := modelInPath(path, "/model/", "/converse"); pathModel != "" {
		kind = bedrockConverse
		model = pathModel
	} else if pathModel := modelInPath(path, "/model/", "/invoke"); pathModel != "" {
		kind = bedrockInvoke
		model = pathModel
	} else {
		return unknownResponse, nil
	}

	provider := helpers.NormalizeHostname(u.Hostname())
	if provider == "" {
		return unknownResponse, nil
	}
	return kind, &protos.AIUsage{
		Provider: strings.Clone(provider),
		Model:    strings.Clone(model),
	}
}

func modelInPath(path, prefix, suffix string) string {
	path, ok := strings.CutSuffix(path, suffix)
	if !ok {
		return ""
	}
	start := strings.LastIndex(path, prefix)
	if start == -1 {
		return ""
	}
	model := path[start+len(prefix):]
	if strings.Contains(model, "/") {
		return ""
	}
	return model
}
